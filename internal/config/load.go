package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	EnvEnvironment         = "CNPJ_LOADER_ENVIRONMENT"
	EnvMySQLHost           = "CNPJ_LOADER_MYSQL_HOST"
	EnvMySQLPort           = "CNPJ_LOADER_MYSQL_PORT"
	EnvMySQLUser           = "CNPJ_LOADER_MYSQL_USER"
	EnvMySQLPassword       = "CNPJ_LOADER_MYSQL_PASSWORD"
	EnvMySQLConnectTimeout = "CNPJ_LOADER_MYSQL_CONNECT_TIMEOUT"
	EnvControlSchema       = "CNPJ_LOADER_CONTROL_SCHEMA"
)

func Load() (Config, error) {
	result := Default()

	if value, exists := os.LookupEnv(EnvEnvironment); exists {
		result.Environment = Environment(value)
	}

	if value, exists := os.LookupEnv(EnvMySQLHost); exists {
		result.MySQL.Host = value
	}

	if value, exists := os.LookupEnv(EnvMySQLPort); exists {
		port, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s deve conter um número inteiro: %w",
				EnvMySQLPort,
				err,
			)
		}

		result.MySQL.Port = port
	}

	if value, exists := os.LookupEnv(EnvMySQLUser); exists {
		result.MySQL.User = value
	}

	if value, exists := os.LookupEnv(EnvMySQLPassword); exists {
		result.MySQL.Password = value
	}

	if value, exists := os.LookupEnv(EnvMySQLConnectTimeout); exists {
		connectTimeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%s deve conter uma duração válida: %w",
				EnvMySQLConnectTimeout,
				err,
			)
		}

		result.MySQL.ConnectTimeout = connectTimeout
	}

	if value, exists := os.LookupEnv(EnvControlSchema); exists {
		result.ControlSchema = value
	}

	return result, nil
}
