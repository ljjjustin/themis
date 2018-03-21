package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "config")

type ThemisConfig struct {
	Debug bool

	// Log configurations
	LogLevel string
	LogFile  string

	// API configurations
	BindHost string
	BindPort int

	Database DatabaseConfig
	Monitors map[string]MonitorConfig
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Username string
	Password string
	Name     string
	Path     string // for sqlite3
}

type MonitorConfig struct {
	Type    string
	Address string
}

func NewConfig(configFile string) *ThemisConfig {

	defaultCfg := NewDefaultConfig()

	if len(configFile) > 0 {
		// override default configurations
		_, err := toml.DecodeFile(configFile, defaultCfg)
		if err != nil {
			plog.Fatalf("Failed to load config file due to %s\n", err)
		}
	}
	return defaultCfg
}

func NewDefaultConfig() *ThemisConfig {
	return &ThemisConfig{
		Debug:    false,
		LogLevel: "INFO",
		LogFile:  "themis.log",
		BindHost: "localhost",
		BindPort: 7878,
		Database: DatabaseConfig{
			Driver:   "sqlite3",
			Path:     "themis.db",
			Host:     "",
			Username: "",
			Password: "",
		},
		Monitors: map[string]MonitorConfig{},
	}
}

func (cfg *ThemisConfig) SetupLogging() {
	plog.Infof("log file path is: %s", cfg.LogFile)

	capnslog.SetGlobalLogLevel(capnslog.INFO)
	if cfg.Debug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	}

	if cfg.LogFile != "" {
		logFile, err := os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			plog.Fatalf("Can't open log file due to %s.", err)
		}
		capnslog.SetFormatter(capnslog.NewPrettyFormatter(logFile, cfg.Debug))
	} else {
		capnslog.SetFormatter(capnslog.NewPrettyFormatter(os.Stderr, cfg.Debug))
	}
}
