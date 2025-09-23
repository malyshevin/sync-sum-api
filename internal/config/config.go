package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Migrate  MigrateConfig
}

type HTTPConfig struct {
	Port int `envconfig:"HTTP_PORT" default:"8080"`
}

type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	Name     string `envconfig:"DB_NAME" default:"syncsum"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func Load(cfg *Config) error {
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}
	return nil
}

type MigrateConfig struct {
	Dir     string `envconfig:"MIGRATIONS_DIR" default:"migrations"`
	Version uint   `envconfig:"MIGRATIONS_VERSION" default:"0"`
}
