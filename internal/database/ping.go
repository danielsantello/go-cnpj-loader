package database

import (
	"context"
	"fmt"
)

type pinger interface {
	PingContext(ctx context.Context) error
}

func PingMySQL(ctx context.Context, connection pinger) error {
	if err := connection.PingContext(ctx); err != nil {
		return fmt.Errorf(
			"não foi possível conectar ao MySQL: %w",
			err,
		)
	}

	return nil
}
