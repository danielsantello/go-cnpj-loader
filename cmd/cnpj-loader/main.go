package main

import (
	"fmt"
	"os"

	"github.com/danielsantello/go-cnpj-loader/internal/cli"
)

func main() {
	command := cli.NewRootCommand(os.Stdout, os.Stderr)

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
