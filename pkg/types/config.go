package types

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Logging     LoggingConfig     `yaml:"logging"`
	Database    DatabaseConfig    `yaml:"database"`
	SSH         SSHConfig         `yaml:"ssh"`
	Security    SecurityConfig    `yaml:"security"`
	System      SystemConfig      `yaml:"system"`
	Application ApplicationConfig `yaml:"application"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
}

type ServerConfig struct {
	ID               string `yaml:"id"`
	Hostname         string `yaml:"hostname"`
	PollingInterval  string `yaml:"polling_interval"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type DatabaseConfig struct {
	Path          string `yaml:"path"`
	MaxSizeMB     int    `yaml:"max_size_mb"`
	RetentionDays int    `yaml:"retention_days"`
}

type SSHConfig struct {
	Enabled      bool     `yaml:"enabled"`
	LogFiles     []string `yaml:"log_files"`
	TrackSessions bool    `yaml:"track_sessions"`
}

type SecurityConfig struct {
	Enabled     bool             `yaml:"enabled"`
	BruteForce  BruteForceConfig `yaml:"brute_force"`
	PortScan    PortScanConfig   `yaml:"port_scan"`
}

type BruteForceConfig struct {
	Threshold      int `yaml:"threshold"`
	WindowMinutes int `yaml:"window_minutes"`
}

type PortScanConfig struct {
	Threshold     int `yaml:"threshold"`
	WindowSeconds int `yaml:"window_seconds"`
}

type SystemConfig struct {
	Enabled    bool     `yaml:"enabled"`
	LogFiles   []string `yaml:"log_files"`
	Journalctl bool     `yaml:"journalctl"`
}

type ApplicationConfig struct {
	Enabled bool                `yaml:"enabled"`
	Nginx   NginxConfig         `yaml:"nginx"`
	Apache  ApacheConfig        `yaml:"apache"`
	PM2     PM2Config           `yaml:"pm2"`
	Custom  []CustomLogConfig  `yaml:"custom"`
}

type NginxConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AccessLog    string `yaml:"access_log"`
	ErrorLog     string `yaml:"error_log"`
	WatchVhosts  bool   `yaml:"watch_vhosts"`
}

type ApacheConfig struct {
	Enabled       bool   `yaml:"enabled"`
	AccessLog     string `yaml:"access_log"`
	ErrorLog      string `yaml:"error_log"`
	RHELAccessLog string `yaml:"rhel_access_log"`
	RHELErrorLog  string `yaml:"rhel_error_log"`
}

type PM2Config struct {
	Enabled     bool   `yaml:"enabled"`
	LogDir      string `yaml:"log_dir"`
	WatchStdout bool   `yaml:"watch_stdout"`
	WatchStderr bool   `yaml:"watch_stderr"`
}

type CustomLogConfig struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type MonitoringConfig struct {
	Realtime   bool `yaml:"realtime"`
	BufferSize int  `yaml:"buffer_size"`
}
