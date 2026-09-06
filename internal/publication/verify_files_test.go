package publication

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyFilesReturnsVerifiedFilesInOriginalOrder(t *testing.T) {
	directory := t.TempDir()

	firstContent := []byte("primeiro arquivo")
	secondContent := []byte("segundo arquivo")

	firstPath := filepath.Join(directory, "Empresas0.zip")
	secondPath := filepath.Join(directory, "Empresas1.zip")

	if err := os.WriteFile(
		firstPath,
		firstContent,
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o primeiro arquivo: %v",
			err,
		)
	}

	if err := os.WriteFile(
		secondPath,
		secondContent,
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o segundo arquivo: %v",
			err,
		)
	}

	files := []ClassifiedFile{
		{
			DatasetCode:    "companies",
			PartNumber:     0,
			SourceName:     "Empresas0.zip",
			SourceLocation: firstPath,
		},
		{
			DatasetCode:    "companies",
			PartNumber:     1,
			SourceName:     "Empresas1.zip",
			SourceLocation: secondPath,
		},
	}

	result, err := VerifyFiles(files)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := []VerifiedFile{
		{
			ClassifiedFile: files[0],
			SizeBytes:      uint64(len(firstContent)),
			SHA256:         sha256.Sum256(firstContent),
		},
		{
			ClassifiedFile: files[1],
			SizeBytes:      uint64(len(secondContent)),
			SHA256:         sha256.Sum256(secondContent),
		},
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"esperava %d arquivos, mas recebeu %d",
			len(expected),
			len(result),
		)
	}

	for index, expectedFile := range expected {
		if result[index] != expectedFile {
			t.Errorf(
				"arquivo %d deveria ser %#v, mas recebeu %#v",
				index,
				expectedFile,
				result[index],
			)
		}
	}
}

func TestVerifyFilesRejectsInvalidCollections(t *testing.T) {
	directory := t.TempDir()

	validFilePath := filepath.Join(directory, "Empresas0.zip")
	if err := os.WriteFile(
		validFilePath,
		[]byte("conteúdo válido"),
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo válido: %v",
			err,
		)
	}

	tests := []struct {
		name            string
		files           []ClassifiedFile
		expectedMessage string
	}{
		{
			name:            "lista vazia",
			files:           nil,
			expectedMessage: "não existem arquivos classificados",
		},
		{
			name: "arquivo inválido após arquivo válido",
			files: []ClassifiedFile{
				{
					DatasetCode:    "companies",
					PartNumber:     0,
					SourceName:     "Empresas0.zip",
					SourceLocation: validFilePath,
				},
				{
					DatasetCode: "companies",
					PartNumber:  1,
					SourceName:  "Empresas1.zip",
					SourceLocation: filepath.Join(
						directory,
						"Empresas1.zip",
					),
				},
			},
			expectedMessage: "não foi possível verificar os arquivos da publicação",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := VerifyFiles(test.files)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != nil {
				t.Errorf(
					"resultado deveria ser nil, mas recebeu: %#v",
					result,
				)
			}

			if !strings.Contains(err.Error(), test.expectedMessage) {
				t.Errorf(
					"erro deveria conter %q, mas recebeu: %v",
					test.expectedMessage,
					err,
				)
			}
		})
	}
}
