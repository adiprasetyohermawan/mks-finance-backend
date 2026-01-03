package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func Load() (Config, error) {
	c := Config{
		AppPort: getenv("API_PORT", "8080"),

		DBHost: getenv("DB_HOST", "127.0.0.1"),
		DBPort: getenv("DB_PORT", "3306"),
		DBUser: getenv("DB_USER", ""),
		DBPass: getenv("DB_PASSWORD", ""),
		DBName: getenv("DB_NAME", ""),
	}

	// minimal validation
	if c.DBUser == "" || c.DBName == "" {
		return c, fmt.Errorf("missing required env: DB_USER and/or DB_NAME")
	}
	return c, nil
}

func (c Config) MySQLDSN() string {
	// parseTime penting untuk DATETIME/DATE
	// loc=UTC aman untuk POC; nanti bisa disesuaikan Asia/Jakarta
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&loc=UTC",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
