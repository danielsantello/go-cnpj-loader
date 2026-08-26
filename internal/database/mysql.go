package database

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/danielsantello/go-cnpj-loader/internal/config"
	"github.com/go-sql-driver/mysql"
)

func OpenMySQL(value config.MySQL) (*sql.DB, error) {
	driverConfig := mysql.NewConfig()
	driverConfig.User = value.User
	driverConfig.Passwd = value.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(
		value.Host,
		strconv.Itoa(value.Port),
	)
	driverConfig.Timeout = value.ConnectTimeout
	driverConfig.ParseTime = true

	connector, err := mysql.NewConnector(driverConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível configurar a conexão com o MySQL: %w",
			err,
		)
	}

	return sql.OpenDB(connector), nil
}
