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

	Storage  StorageConfig
	Monitors map[string]MonitorConfig
}

type StorageConfig struct {
	// Driver name
	Driver string

	// Connect URLs
	Connection string
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
		Storage: StorageConfig{
			Driver:     "mysql",
			Connection: "mysql://localhost/themis?charset=utf8",
		},
		Monitors: map[string]MonitorConfig{
			"ceph": MonitorConfig{
				Type:    "serf",
				Address: "http://127.0.0.1:7373",
			},
		},
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
