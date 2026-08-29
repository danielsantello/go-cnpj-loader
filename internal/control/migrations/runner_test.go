package migrations

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestMySQLErrorDetailsReturnsDriverMetadata(t *testing.T) {
	driverError := &mysql.MySQLError{
		Number:   3819,
		SQLState: [5]byte{'H', 'Y', '0', '0', '0'},
		Message:  "check constraint is violated",
	}

	wrappedError := fmt.Errorf(
		"não foi possível executar comando: %w",
		driverError,
	)

	errorCode, sqlState := mysqlErrorDetails(wrappedError)

	if !errorCode.Valid {
		t.Fatal("código de erro do MySQL deveria ser válido")
	}

	if errorCode.Int64 != 3819 {
		t.Errorf(
			"código de erro deveria ser %d, mas recebeu %d",
			3819,
			errorCode.Int64,
		)
	}

	if !sqlState.Valid {
		t.Fatal("SQL state deveria ser válido")
	}

	if sqlState.String != "HY000" {
		t.Errorf(
			"SQL state deveria ser %q, mas recebeu %q",
			"HY000",
			sqlState.String,
		)
	}
}

func TestMySQLErrorDetailsReturnsNullForOtherErrors(t *testing.T) {
	errorCode, sqlState := mysqlErrorDetails(
		errors.New("erro sem metadata do MySQL"),
	)

	if errorCode.Valid {
		t.Errorf(
			"código de erro não deveria ser válido, mas recebeu %d",
			errorCode.Int64,
		)
	}

	if sqlState.Valid {
		t.Errorf(
			"SQL state não deveria ser válido, mas recebeu %q",
			sqlState.String,
		)
	}
}
