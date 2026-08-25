package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandDisplaysHelp(t *testing.T) {
	var output bytes.Buffer

	command := NewRootCommand(&output, &output)
	command.SetArgs([]string{})

	err := command.Execute()
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	result := output.String()

	expectedContents := []string{
		"cnpj-loader é um carregador versionado",
		"Usage:",
		"cnpj-loader [flags]",
		"--help",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("a saída deveria conter %q, mas foi:\n%s", expected, result)
		}
	}
}
