package config

import (
	"os"
	"strconv"
)

// Config アプリケーション全体の設定を管理する構造体
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
}

// ServerConfig サーバー関連の設定
type ServerConfig struct {
	Port string `json:"port"`
	Mode string `json:"mode"` // gin.DebugMode, gin.ReleaseMode, gin.TestMode
}

// DatabaseConfig データベース接続設定
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
}

// JWTConfig JWT認証設定
type JWTConfig struct {
	SecretKey      string `json:"secret_key"`
	ExpireHours    int    `json:"expire_hours"`
	RefreshHours   int    `json:"refresh_hours"`
	CookieName     string `json:"cookie_name"`
	CookieSecure   bool   `json:"cookie_secure"`
	CookieHTTPOnly bool   `json:"cookie_http_only"`
	CookieSameSite string `json:"cookie_same_site"`
}

// Load 環境変数から設定を読み込みます
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "simple_kanban"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey:      getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
			ExpireHours:    getEnvAsInt("JWT_EXPIRE_HOURS", 24),
			RefreshHours:   getEnvAsInt("JWT_REFRESH_HOURS", 168), // 1週間
			CookieName:     getEnv("JWT_COOKIE_NAME", "auth_token"),
			CookieSecure:   getEnvAsBool("JWT_COOKIE_SECURE", false),
			CookieHTTPOnly: getEnvAsBool("JWT_COOKIE_HTTP_ONLY", true),
			CookieSameSite: getEnv("JWT_COOKIE_SAME_SITE", "Lax"),
		},
	}
}

// getEnv 環境変数を取得、存在しない場合はデフォルト値を返す
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt 環境変数を整数として取得
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsBool 環境変数をブール値として取得
func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// GetDSN データベース接続文字列を生成
func (c *Config) GetDSN() string {
	return "host=" + c.Database.Host +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" port=" + strconv.Itoa(c.Database.Port) +
		" sslmode=" + c.Database.SSLMode +
		" TimeZone=UTC"
}
