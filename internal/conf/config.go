package conf

import (
	"errors"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddrress           string `env:"RUN_ADDRESS"`
	DatabaseURI           string `env:"DATABASE_URI"`
	AccrualSystemAddrress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey             string `env:"SECRET_KEY"`
}

const (
	DefaultRunAddress           = ":8080"
	DefaultDatabaseURI          = ""
	DefaultAccrualSystemAddress = "http://localhost:8081"
	DefaultSecretKey            = "secret"
)

func ParseConfig() (*Config, error) {
	runAddress := flag.String("a", DefaultRunAddress, "Адрес сервера (в формате host:port)")
	databaseURI := flag.String("d", DefaultDatabaseURI, "Адрес подключения к базе данных (URI)")
	accrualSystemAddress := flag.String("r", DefaultAccrualSystemAddress, "Адрес системы расчета начислений (в формате host:port)")
	secretKey := flag.String("s", DefaultSecretKey, "Секретный ключ для аутентификации")

	flag.Parse()

	cfg := &Config{
		RunAddrress:           *runAddress,
		DatabaseURI:           *databaseURI,
		AccrualSystemAddrress: *accrualSystemAddress,
		SecretKey:             *secretKey,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации: %w", err)
	}

	// проверка конфига
	if cfg.RunAddrress == "" {
		return nil, errors.New("адрес сервера не может быть пустым")
	}

	if cfg.AccrualSystemAddrress == "" {
		return nil, errors.New("адрес системы расчета начислений не может быть пустым")
	}

	return cfg, nil
}
