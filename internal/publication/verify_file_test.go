package publication

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyFileReturnsSizeAndSHA256(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "Empresas0.zip")
	content := []byte("conteúdo do arquivo de teste")

	if err := os.WriteFile(
		filePath,
		content,
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo de teste: %v",
			err,
		)
	}

	classifiedFile := ClassifiedFile{
		DatasetCode:    "companies",
		PartNumber:     0,
		SourceName:     "Empresas0.zip",
		SourceLocation: filePath,
	}

	result, err := VerifyFile(classifiedFile)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := VerifiedFile{
		ClassifiedFile: classifiedFile,
		SizeBytes:      uint64(len(content)),
		SHA256:         sha256.Sum256(content),
	}

	if result != expected {
		t.Errorf(
			"arquivo verificado deveria ser %#v, mas recebeu %#v",
			expected,
			result,
		)
	}
}

func TestVerifyFileRejectsInvalidFiles(t *testing.T) {
	directory := t.TempDir()

	regularFilePath := filepath.Join(directory, "Empresas0.zip")
	if err := os.WriteFile(
		regularFilePath,
		[]byte("conteúdo de teste"),
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo regular: %v",
			err,
		)
	}

	emptyFilePath := filepath.Join(directory, "Vazio.zip")
	if err := os.WriteFile(
		emptyFilePath,
		nil,
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo vazio: %v",
			err,
		)
	}

	tests := []struct {
		name            string
		file            ClassifiedFile
		expectedMessage string
	}{
		{
			name: "nome vazio",
			file: ClassifiedFile{
				SourceName:     "",
				SourceLocation: regularFilePath,
			},
			expectedMessage: "nome do arquivo classificado é obrigatório",
		},
		{
			name: "localização vazia",
			file: ClassifiedFile{
				SourceName:     "Empresas0.zip",
				SourceLocation: "",
			},
			expectedMessage: "localização do arquivo classificado é obrigatória",
		},
		{
			name: "caminho relativo",
			file: ClassifiedFile{
				SourceName:     "Empresas0.zip",
				SourceLocation: "Empresas0.zip",
			},
			expectedMessage: "deve ser absoluta",
		},
		{
			name: "arquivo inexistente",
			file: ClassifiedFile{
				SourceName: "Inexistente.zip",
				SourceLocation: filepath.Join(
					directory,
					"Inexistente.zip",
				),
			},
			expectedMessage: "não foi possível abrir",
		},
		{
			name: "diretório no lugar do arquivo",
			file: ClassifiedFile{
				SourceName:     "diretorio",
				SourceLocation: directory,
			},
			expectedMessage: "não é um arquivo regular",
		},
		{
			name: "arquivo vazio",
			file: ClassifiedFile{
				SourceName:     "Vazio.zip",
				SourceLocation: emptyFilePath,
			},
			expectedMessage: "está vazio",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := VerifyFile(test.file)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != (VerifiedFile{}) {
				t.Errorf(
					"resultado deveria estar vazio, mas recebeu: %#v",
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
