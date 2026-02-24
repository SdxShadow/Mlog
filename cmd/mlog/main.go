package main

import (
	"fmt"
	"os"

	"github.com/mlog/mlog/internal/config"
	"github.com/mlog/mlog/internal/db"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mlog",
	Short: "Mlog - Linux server monitoring and logging tool",
	Long:  `Mlog captures SSH connections, security events, and system logs from Linux servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Mlog CLI - Use --help for usage")
	},
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query events from the database",
	Run:   runQuery,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor logs in real-time",
	Run:   runMonitor,
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
	Run:   runStats,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Run:   runConfig,
}

func main() {
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(configCmd)

	queryCmd.Flags().StringP("type", "t", "", "Filter by event type")
	queryCmd.Flags().StringP("ip", "i", "", "Filter by source IP")
	queryCmd.Flags().StringP("user", "u", "", "Filter by username")
	queryCmd.Flags().StringP("severity", "s", "", "Filter by severity")
	queryCmd.Flags().Int("limit", 100, "Limit results")

	monitorCmd.Flags().BoolP("ssh", "s", false, "Only show SSH events")
	monitorCmd.Flags().BoolP("security", "S", false, "Only show security events")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initDB() error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	return db.Init(cfg.Database.Path)
}

func runQuery(cmd *cobra.Command, args []string) {
	if err := initDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init DB: %v\\n", err)
		os.Exit(1)
	}
	defer db.Close()

	eventType, _ := cmd.Flags().GetString("type")
	sourceIP, _ := cmd.Flags().GetString("ip")
	username, _ := cmd.Flags().GetString("user")
	severity, _ := cmd.Flags().GetString("severity")
	limit, _ := cmd.Flags().GetInt("limit")

	q := &db.EventQuery{
		EventType: eventType,
		SourceIP:  sourceIP,
		Username:  username,
		Severity:  severity,
		Limit:     limit,
	}

	events, err := db.QueryEvents(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d events:\\n", len(events))
	for _, e := range events {
		fmt.Printf("[%s] %s - %s - %s\\n", e.Timestamp.Format("2006-01-02 15:04:05"), e.EventType, e.Severity, e.Message)
	}
}

func runMonitor(cmd *cobra.Command, args []string) {
	fmt.Println("Real-time monitoring - not yet implemented")
}

func runStats(cmd *cobra.Command, args []string) {
	fmt.Println("Statistics - not yet implemented")
}

func runConfig(cmd *cobra.Command, args []string) {
	fmt.Println("Configuration - not yet implemented")
}
