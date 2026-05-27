package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration settings
type Config struct {
	DefaultEnv    string            `mapstructure:"defaultEnv"`
	Parallelism   int               `mapstructure:"parallelism"`
	CacheDir      string            `mapstructure:"cacheDir"`
	AuditDbPath   string            `mapstructure:"auditDbPath"`
	AuditCacheTTL string            `mapstructure:"auditCacheTTL"`
	NotifyOnCVE   bool              `mapstructure:"notifyOnCVE"`
	WebhookURL    string            `mapstructure:"webhookURL"`
	LicensePolicy string            `mapstructure:"licensePolicy"`
	Aliases       map[string]string `mapstructure:"aliases"`
}

var AppConfig *Config

// ExpandTilde replaces the tilde character at the start of a path with the user's home directory.
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// LoadConfig initializes Viper and reads config.toml, setting sensible defaults if it doesn't exist.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".gopack")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, "config.toml")

	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")

	// Set defaults
	viper.SetDefault("defaultEnv", "dev")
	viper.SetDefault("parallelism", 8)
	viper.SetDefault("cacheDir", filepath.Join(configDir, "store"))
	viper.SetDefault("auditDbPath", filepath.Join(configDir, "osv.db"))
	viper.SetDefault("auditCacheTTL", "1h")
	viper.SetDefault("notifyOnCVE", true)
	viper.SetDefault("webhookURL", "")
	viper.SetDefault("licensePolicy", "permissive")
	viper.SetDefault("aliases", map[string]string{
		"dev":   "install --env dev",
		"fresh": "cache clean && install",
	})

	// If config file doesn't exist, write defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := viper.WriteConfigAs(configFile); err != nil {
			return nil, err
		}
	} else {
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	// Expand paths
	conf.CacheDir = ExpandTilde(conf.CacheDir)
	conf.AuditDbPath = ExpandTilde(conf.AuditDbPath)

	AppConfig = &conf
	return &conf, nil
}
