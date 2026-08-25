package cli

import (
	"bytes"
	"testing"

	"github.com/danielsantello/go-cnpj-loader/internal/buildinfo"
)

func TestVersionCommandDisplaysBuildInformation(t *testing.T) {
	var output bytes.Buffer

	info := buildinfo.Info{
		Version:   "v0.1.0",
		Commit:    "a1b2c3d",
		BuildDate: "2026-08-25T18:00:00Z",
		GoVersion: "go1.27.0",
	}

	command := newVersionCommand(info)
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{})

	err := command.Execute()
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := "" +
		"Versão: v0.1.0\n" +
		"Commit: a1b2c3d\n" +
		"Compilado em: 2026-08-25T18:00:00Z\n" +
		"Go: go1.27.0\n"

	if output.String() != expected {
		t.Errorf("saída inesperada:\n%s", output.String())
	}
}
