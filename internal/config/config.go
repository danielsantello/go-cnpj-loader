package config

import "time"

const (
	DefaultMySQLHost      = "127.0.0.1"
	DefaultMySQLPort      = 3306
	DefaultControlSchema  = "cnpj_loader_control"
	DefaultConnectTimeout = 5 * time.Second
)

type Config struct {
	MySQL         MySQL
	ControlSchema string
}

type MySQL struct {
	Host           string
	Port           int
	User           string
	Password       string
	ConnectTimeout time.Duration
}

func Default() Config {
	return Config{
		MySQL: MySQL{
			Host:           DefaultMySQLHost,
			Port:           DefaultMySQLPort,
			ConnectTimeout: DefaultConnectTimeout,
		},
		ControlSchema: DefaultControlSchema,
	}
}
