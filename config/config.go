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

	Fence FenceConfig

	Openstack OpenstackConfig

	CatKeeper CatkeeperConfig
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

type FenceConfig struct {
	DisableFenceOps bool
}

type OpenstackConfig struct {
	AuthURL     string
	Username    string
	Password    string
	ProjectName string
	DomainName  string
	RegionName  string
}

type CatkeeperConfig struct {
	Url	 string
	Username string
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
		Fence: FenceConfig{
			DisableFenceOps: false,
		},
		Openstack: OpenstackConfig{
			AuthURL:     "http://localhost:5000",
			Username:    "admin",
			Password:    "secretxx",
			ProjectName: "admin",
			DomainName:  "default",
			RegionName:  "RegionOne",
		},
		CatKeeper: CatkeeperConfig{
			Url:		"http://127.0.0.1",
			Username:	"admin",
		},
	}
}

func (cfg *ThemisConfig) SetupLogging() {
	if cfg.Debug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	} else {
		capnslog.SetGlobalLogLevel(capnslog.INFO)
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
