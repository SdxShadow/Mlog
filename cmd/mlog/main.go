package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SdxShadow/Mlog/internal/config"
	"github.com/SdxShadow/Mlog/internal/db"
	"github.com/SdxShadow/Mlog/internal/monitor"
	"github.com/SdxShadow/Mlog/pkg/types"
	"github.com/spf13/cobra"
)

func forkExec(name string, args []string) (int, error) {
	pid, err := syscall.ForkExec(name, args, &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
	})
	return pid, err
}

var version = "v0.1.4"

var rootCmd = &cobra.Command{
	Use:   "mlog",
	Short: "Mlog - Linux server monitoring and logging",
	Long:  `Mlog monitors SSH, Nginx, Apache, PM2 logs and system stats`,
	Version: version,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as daemon to collect logs",
	Run:   runServe,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service for auto-start",
	Run:   runInstall,
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Live dashboard with logs and system stats",
	Run:   runDashboard,
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query events from database",
	Run:   runQuery,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running mlog daemon",
	Run:   runStop,
}

func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(stopCmd)

	serveCmd.Flags().StringP("config", "c", "", "Config file path")
	serveCmd.Flags().BoolP("daemon", "d", false, "Run in background")
	dashboardCmd.Flags().StringP("config", "c", "", "Config file path")
	queryCmd.Flags().StringP("type", "t", "", "Event type filter")
	queryCmd.Flags().StringP("ip", "i", "", "Source IP filter")
	queryCmd.Flags().Int("limit", 50, "Result limit")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServe(cmd *cobra.Command, args []string) {
	daemon, _ := cmd.Flags().GetBool("daemon")
	configPath, _ := cmd.Flags().GetString("config")

	// Run in background
	if daemon {
		execPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get executable: %v\n", err)
			os.Exit(1)
		}
		args := []string{"serve"}
		if configPath != "" {
			args = append(args, "-c", configPath)
		}
		pid, err := forkExec(execPath, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to daemonize: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Mlog started in background (PID: %d)\n", pid)
		return
	}

	cfg, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	if os.Geteuid() != 0 {
		fmt.Println("Warning: Not running as root. Some logs may not be accessible.")
	}

	if err := db.Init(cfg.Database.Path); err != nil {
		fmt.Fprintf(os.Stderr, "DB error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Printf("Mlog serving on: %s\n", cfg.Server.ID)

	w := monitor.NewWatcher(cfg.Server.ID)

	for _, f := range cfg.SSH.LogFiles {
		if exists(f) {
			w.AddPath(f)
		}
	}

	if cfg.Application.Nginx.Enabled {
		w.AddPath(cfg.Application.Nginx.AccessLog)
		w.AddPath(cfg.Application.Nginx.ErrorLog)
	}

	if cfg.Application.Apache.Enabled {
		w.AddPath(cfg.Application.Apache.AccessLog)
		w.AddPath(cfg.Application.Apache.ErrorLog)
	}

	if cfg.Application.PM2.Enabled {
		expandPath(&cfg.Application.PM2.LogDir)
	}

	if err := w.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		os.Exit(1)
	}
	defer w.Stop()

	fmt.Println("Monitoring started. Press Ctrl+C to stop.")
	select {}
}

func runDashboard(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config")

	cfg, err := loadOrCreateConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	if err := db.Init(cfg.Database.Path); err != nil {
		fmt.Fprintf(os.Stderr, "DB error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dash := &Dashboard{maxLines: 30}
	dash.Start()
}

func runQuery(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config")
	cfg, _ := loadOrCreateConfig(configPath)
	if cfg == nil {
		cfg = defaultConfig()
	}

	db.Init(cfg.Database.Path)
	defer db.Close()

	eventType, _ := cmd.Flags().GetString("type")
	ip, _ := cmd.Flags().GetString("ip")
	limit, _ := cmd.Flags().GetInt("limit")

	events, err := db.QueryEvents(&db.EventQuery{
		EventType: eventType,
		SourceIP:  ip,
		Limit:     limit,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		return
	}

	for _, e := range events {
		fmt.Printf("[%s] %-20s %s\n", e.Timestamp.Format("15:04:05"), e.EventType, e.Message)
	}
}

func runInstall(cmd *cobra.Command, args []string) {
	// Get mlog binary path
	binPath, err := exec.LookPath("mlog")
	if err != nil {
		fmt.Println("Error: mlog not found in PATH. Install with: go install github.com/SdxShadow/Mlog/cmd/mlog@v0.1.4")
		os.Exit(1)
	}

	service := `[Unit]
Description=Mlog - Linux Server Monitoring
After=network.target

[Service]
Type=simple
ExecStart=` + binPath + ` serve -d
ExecStop=` + binPath + ` stop
Restart=always
RestartSec=10
User=root

[Install]
WantedBy=multi-user.target
`

	// Write service file
	err = os.WriteFile("/etc/systemd/system/mlog.service", []byte(service), 0644)
	if err != nil {
		fmt.Printf("Error writing service file: %v\n", err)
		os.Exit(1)
	}

	// Reload systemd
	cmdExec := exec.Command("systemctl", "daemon-reload")
	cmdExec.Run()

	fmt.Println("Mlog service installed!")
	fmt.Println("Commands:")
	fmt.Println("  sudo systemctl start mlog   # Start")
	fmt.Println("  sudo systemctl stop mlog    # Stop")
	fmt.Println("  sudo systemctl status mlog  # Status")
}

func runStop(cmd *cobra.Command, args []string) {
	// Find mlog process using pgrep
	cmdExec := exec.Command("pgrep", "-f", "mlog serve")
	output, err := cmdExec.Output()
	if err != nil {
		fmt.Println("No mlog daemon running")
		return
	}

	var pids []int
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var pid int
		fmt.Sscanf(line, "%d", &pid)
		if pid > 0 && pid != os.Getpid() {
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		fmt.Println("No mlog daemon running")
		return
	}

	for _, pid := range pids {
		err := syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			fmt.Printf("Failed to stop process %d: %v\n", pid, err)
		} else {
			fmt.Printf("Stopped mlog daemon (PID: %d)\n", pid)
		}
	}
}

func loadOrCreateConfig(path string) (*types.Config, error) {
	if exists(path) {
		return config.Load(path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return defaultConfig(), nil
	}

	cfg := defaultConfig()
	if err := config.Save(path, cfg); err != nil {
		return cfg, nil
	}

	fmt.Printf("Created default config: %s\n", path)
	return cfg, nil
}

func defaultConfig() *types.Config {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/tmp"
	}
	return &types.Config{
		Server: types.ServerConfig{
			ID:              getHostname(),
			PollingInterval: "1s",
		},
		Database: types.DatabaseConfig{
			Path:          home + "/.mlog/mlog.db",
			MaxSizeMB:     1000,
			RetentionDays: 90,
		},
		SSH: types.SSHConfig{
			Enabled:       true,
			LogFiles:      []string{"/var/log/auth.log", "/var/log/secure"},
			TrackSessions: true,
		},
		Security: types.SecurityConfig{
			Enabled: true,
			BruteForce: types.BruteForceConfig{
				Threshold:      5,
				WindowMinutes: 5,
			},
		},
		Application: types.ApplicationConfig{
			Enabled: true,
			Nginx: types.NginxConfig{
				Enabled:     true,
				AccessLog:  "/var/log/nginx/access.log",
				ErrorLog:   "/var/log/nginx/error.log",
			},
			Apache: types.ApacheConfig{
				Enabled:    true,
				AccessLog:  "/var/log/apache2/access.log",
				ErrorLog:   "/var/log/apache2/error.log",
			},
			PM2: types.PM2Config{
				Enabled: true,
				LogDir:  os.ExpandEnv("$HOME/.pm2/logs"),
			},
		},
	}
}

func getHostname() string {
	h, _ := os.Hostname()
	if h == "" {
		return "localhost"
	}
	return h
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func expandPath(p *string) {
	*p = os.ExpandEnv(*p)
}

// Dashboard for live view
type Dashboard struct {
	events   []*types.Event
	mu       sync.RWMutex
	stopCh   chan bool
	maxLines int
}

func (d *Dashboard) Start() {
	go d.pollEvents()
	d.render()
	<-d.stopCh
}

func (d *Dashboard) pollEvents() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			events, _ := db.QueryEvents(&db.EventQuery{Limit: d.maxLines})
			d.mu.Lock()
			d.events = events
			d.mu.Unlock()
			d.render()
		case <-d.stopCh:
			return
		}
	}
}

func (d *Dashboard) render() {
	d.mu.RLock()
	events := d.events
	d.mu.RUnlock()

	fmt.Print("\033[2J\033[H")

	fmt.Println("\033[1;34m┌────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    MLOG MONITOR                               │")
	fmt.Println("└────────────────────────────────────────────────────────────────┘\033[0m")

	fmt.Println("\033[1;33m┌────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ SYSTEM STATUS                                                │")
	fmt.Println("├────────────────────────────────────────────────────────────────┤")

	out, _ := exec.Command("sh", "-c", "uptime && free -h && df -h / | tail -1").Output()
	fmt.Printf("\033[90m%s\033[0m", out)
	fmt.Println("└────────────────────────────────────────────────────────────────┘\033[0m")

	fmt.Println("\033[1;36m┌────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ LIVE LOGS                                                    │")
	fmt.Println("├────────────────────────────────────────────────────────────────┤\033[0m")

	for _, e := range events {
		color := getColor(e.EventType)
		fmt.Printf("%s[%s] %-15s %s\033[0m\n",
			color,
			e.Timestamp.Format("15:04:05"),
			e.SourceIP,
			trunc(e.Message, 50))
	}

	if len(events) == 0 {
		fmt.Println("\033[90m  Waiting for logs...\033[0m")
	}

	fmt.Println("\033[1;36m└────────────────────────────────────────────────────────────────┘\033[0m")
	fmt.Println("\033[90m  Ctrl+C to exit\033[0m")
}

func getColor(t types.EventType) string {
	s := string(t)
	if strings.Contains(s, "CONNECTED") || strings.Contains(s, "START") {
		return "\033[32m"
	}
	if strings.Contains(s, "FAILED") || strings.Contains(s, "ERROR") || strings.Contains(s, "CRASH") {
		return "\033[31m"
	}
	if strings.Contains(s, "WARNING") || strings.Contains(s, "STOP") {
		return "\033[33m"
	}
	return "\033[37m"
}

func trunc(s string, l int) string {
	if len(s) > l {
		return s[:l-3] + "..."
	}
	return s
}
