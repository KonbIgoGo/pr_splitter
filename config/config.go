package config

import "os"

type Config struct {
	URL      string
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	DB       string `env:"DB"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	MaxConn  string `env:"MAX_CONN"`
}

func New() (*Config, error) {
	cfg := &Config{}

	cfg.Host = os.Getenv("HOST")
	cfg.Port = os.Getenv("PORT")
	cfg.DB = os.Getenv("DB")
	cfg.User = os.Getenv("USER")
	cfg.Password = os.Getenv("PASSWORD")
	cfg.MaxConn = os.Getenv("MAX_CONN")

	return cfg, nil
}
