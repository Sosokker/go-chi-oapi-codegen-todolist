package config

import (
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Cache    CacheConfig
	Storage  StorageConfig
	Frontend FrontendConfig
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
	BasePath     string        `mapstructure:"basePath"`
}

type StorageConfig struct {
	Type  string             `mapstructure:"type"` // "local", "gcs"
	Local LocalStorageConfig `mapstructure:"local"`
	GCS   GCSStorageConfig   `mapstructure:"gcs"`
}

type LocalStorageConfig struct {
	Path string `mapstructure:"path"`
}

type GCSStorageConfig struct {
	BucketName      string `mapstructure:"bucketName"`
	CredentialsFile string `mapstructure:"credentialsFile"`
	BaseDir         string `mapstructure:"baseDir"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	ExpiryMinutes  int    `mapstructure:"expiryMinutes"`
	CookieName     string `mapstructure:"cookieName"`
	CookieDomain   string `mapstructure:"cookieDomain"`
	CookiePath     string `mapstructure:"cookiePath"`
	CookieSecure   bool   `mapstructure:"cookieSecure"`
	CookieHttpOnly bool   `mapstructure:"cookieHttpOnly"`
	CookieSameSite string `mapstructure:"cookieSameSite"` // None, Lax, Strict
}

type GoogleOAuthConfig struct {
	ClientID     string   `mapstructure:"clientId"`
	ClientSecret string   `mapstructure:"clientSecret"`
	RedirectURL  string   `mapstructure:"redirectUrl"`
	Scopes       []string `mapstructure:"scopes"`
	StateSecret  string   `mapstructure:"stateSecret"` // For signing state cookie
}

type OAuthConfig struct {
	Google GoogleOAuthConfig `mapstructure:"google"`
}

type CacheConfig struct {
	DefaultExpiration time.Duration `mapstructure:"defaultExpiration"`
	CleanupInterval   time.Duration `mapstructure:"cleanupInterval"`
}

type FrontendConfig struct {
	Url string `mapstructure:"url"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // or viper.SetConfigType("YAML")
	viper.AddConfigPath(path)     // optionally look for config in the working directory or specified path
	viper.AddConfigPath(".")      // look for config in the working directory

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", 15*time.Second)
	viper.SetDefault("server.writeTimeout", 15*time.Second)
	viper.SetDefault("server.idleTimeout", 60*time.Second)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("jwt.expiryMinutes", 60)
	viper.SetDefault("jwt.cookieName", "jwt_token")
	viper.SetDefault("jwt.cookieHttpOnly", true)
	viper.SetDefault("jwt.cookieSameSite", "Lax")
	viper.SetDefault("cache.defaultExpiration", 5*time.Minute)
	viper.SetDefault("cache.cleanupInterval", 10*time.Minute)
	viper.SetDefault("storage.type", "local") // Default to local storage
	viper.SetDefault("storage.local.path", "./uploads")
	viper.SetDefault("frontend.url", "http://localhost:3000")

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("Error reading config file, using defaults/env vars", "error", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.JWT.Secret == "" || strings.Contains(cfg.JWT.Secret, "unsafe") {
		slog.Warn("JWT_SECRET environment variable not set or is using unsafe default. THIS IS INSECURE FOR PRODUCTION.")
	}
	if cfg.OAuth.Google.ClientID == "" || strings.Contains(cfg.OAuth.Google.ClientID, "_ENV") {
		slog.Warn("OAUTH_GOOGLE_CLIENTID environment variable not set or is using placeholder.")
	}
	if cfg.OAuth.Google.ClientSecret == "" || strings.Contains(cfg.OAuth.Google.ClientSecret, "_ENV") {
		slog.Warn("OAUTH_GOOGLE_CLIENTSECRET environment variable not set or is using placeholder.")
	}
	if cfg.OAuth.Google.StateSecret == "" || strings.Contains(cfg.OAuth.Google.StateSecret, "unsafe") {
		slog.Warn("OAUTH_GOOGLE_STATESECRET environment variable not set or is using unsafe default. THIS IS INSECURE FOR PRODUCTION.")
	}

	return &cfg, nil
}
