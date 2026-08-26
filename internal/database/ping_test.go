package database

import (
	"context"
	"errors"
	"testing"
)

type pingerStub struct {
	err error
}

func (stub pingerStub) PingContext(context.Context) error {
	return stub.err
}

func TestPingMySQLReturnsNilWhenConnectionSucceeds(t *testing.T) {
	connection := pingerStub{}

	err := PingMySQL(context.Background(), connection)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}
}

func TestPingMySQLWrapsConnectionError(t *testing.T) {
	connectionError := errors.New("connection refused")
	connection := pingerStub{
		err: connectionError,
	}

	err := PingMySQL(context.Background(), connection)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if !errors.Is(err, connectionError) {
		t.Errorf(
			"erro deveria preservar a causa original, mas recebeu: %v",
			err,
		)
	}
}
