package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port                  string
	DBDriver              string
	DBDSN                 string
	SQLitePath            string
	JWTSecret             string
	UploadDir             string
	StorageBackend        string
	AliyunEndpoint        string
	AliyunAccessKeyID     string
	AliyunAccessKeySecret string
	AliyunBucketName      string
	AliyunSignName        string
	AliyunTemplateCode    string
}

func Load() *Config {
	filePath := os.Getenv("CONFIG_FILE")
	if filePath == "" {
		filePath = "config.json"
	}

	var fileCfg *Config
	if exists(filePath) {
		cfg, err := loadFromFile(filePath)
		if err == nil {
			fileCfg = cfg
		} else {
			fmt.Fprintf(os.Stderr, "warn: load config file %s failed: %v\n", filePath, err)
		}
	}

	cfg := &Config{}

	// Resolve with priority: env > file > default
	cfg.Port = firstNonEmpty(os.Getenv("PORT"), fileCfgValue(fileCfg, func(c *Config) string { return c.Port }), "9527")
	cfg.DBDriver = firstNonEmpty(os.Getenv("DB_DRIVER"), fileCfgValue(fileCfg, func(c *Config) string { return c.DBDriver }), "mysql")
	cfg.DBDSN = firstNonEmpty(os.Getenv("DB_DSN"), fileCfgValue(fileCfg, func(c *Config) string { return c.DBDSN }), "chirp:test12345@tcp(127.0.0.1:3306)/chirp?parseTime=true&loc=Local")
	cfg.SQLitePath = firstNonEmpty(os.Getenv("SQLITE_PATH"), fileCfgValue(fileCfg, func(c *Config) string { return c.SQLitePath }), "chirp.db")
	cfg.JWTSecret = firstNonEmpty(os.Getenv("JWT_SECRET"), fileCfgValue(fileCfg, func(c *Config) string { return c.JWTSecret }), "default_secret")
	cfg.UploadDir = firstNonEmpty(os.Getenv("UPLOAD_DIR"), fileCfgValue(fileCfg, func(c *Config) string { return c.UploadDir }), "uploads")
	cfg.StorageBackend = firstNonEmpty(os.Getenv("STORAGE_BACKEND"), fileCfgValue(fileCfg, func(c *Config) string { return c.StorageBackend }), "local")
	cfg.AliyunEndpoint = firstNonEmpty(os.Getenv("ALIYUN_OSS_ENDPOINT"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunEndpoint }), "")
	cfg.AliyunAccessKeyID = firstNonEmpty(os.Getenv("ALIYUN_ACCESS_KEY"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunAccessKeyID }), "")
	cfg.AliyunAccessKeySecret = firstNonEmpty(os.Getenv("ALIYUN_ACCESS_SECRET"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunAccessKeySecret }), "")
	cfg.AliyunBucketName = firstNonEmpty(os.Getenv("ALIYUN_OSS_BUCKET"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunBucketName }), "")
	cfg.AliyunSignName = firstNonEmpty(os.Getenv("ALIYUN_SIGN_NAME"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunSignName }), "")
	cfg.AliyunTemplateCode = firstNonEmpty(os.Getenv("ALIYUN_TEMPLATE_CODE"), fileCfgValue(fileCfg, func(c *Config) string { return c.AliyunTemplateCode }), "")

	return cfg
}

func loadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func exists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// helper to safely read from optional cfg
func fileCfgValue(cfg *Config, getter func(*Config) string) string {
	if cfg == nil {
		return ""
	}
	return getter(cfg)
}
