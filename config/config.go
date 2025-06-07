package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Env  string `envconfig:"ENV" default:"dev"`
	Port string `envconfig:"PORT" default:"80"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBCustomer string `envconfig:"DB_USER" default:"postgres"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName     string `envconfig:"DB_NAME" default:"postgres"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"debug"`
}

func Get() (*Config, error) {
	c := &Config{}
	err := envconfig.Process("", c)
	if err != nil {
		return &Config{}, err
	}

	return c, nil
}
