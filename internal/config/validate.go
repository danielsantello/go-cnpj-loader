package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var schemaNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func Validate(value Config) error {
	var problems []error

	switch value.Environment {
	case EnvironmentDevelopment:
	case EnvironmentBenchmark:
	case EnvironmentProduction:
	default:
		problems = append(
			problems,
			fmt.Errorf(
				"%s possui ambiente inválido: %q",
				EnvEnvironment,
				value.Environment,
			),
		)
	}

	if strings.TrimSpace(value.MySQL.Host) == "" {
		problems = append(
			problems,
			fmt.Errorf("%s é obrigatória", EnvMySQLHost),
		)
	}

	if value.MySQL.Port < 1 || value.MySQL.Port > 65535 {
		problems = append(
			problems,
			fmt.Errorf(
				"%s deve estar entre 1 e 65535",
				EnvMySQLPort,
			),
		)
	}

	if strings.TrimSpace(value.MySQL.User) == "" {
		problems = append(
			problems,
			fmt.Errorf("%s é obrigatória", EnvMySQLUser),
		)
	}

	if value.MySQL.Password == "" {
		problems = append(
			problems,
			fmt.Errorf("%s é obrigatória", EnvMySQLPassword),
		)
	}

	if value.MySQL.ConnectTimeout <= 0 {
		problems = append(
			problems,
			fmt.Errorf(
				"%s deve ser maior que zero",
				EnvMySQLConnectTimeout,
			),
		)
	}

	if !schemaNamePattern.MatchString(value.ControlSchema) {
		problems = append(
			problems,
			fmt.Errorf(
				"%s deve ser um identificador MySQL válido em letras minúsculas",
				EnvControlSchema,
			),
		)
	}

	return errors.Join(problems...)
}
