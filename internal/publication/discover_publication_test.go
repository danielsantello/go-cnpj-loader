package publication

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverDirectoryPublicationReturnsValidatedFiles(t *testing.T) {
	directory := t.TempDir()
	content := []byte("conteúdo de teste")

	fileNames := []string{
		"Paises.zip",
		"Empresas1.zip",
		"Empresas0.zip",
	}

	for _, fileName := range fileNames {
		filePath := filepath.Join(directory, fileName)

		if err := os.WriteFile(
			filePath,
			content,
			0o600,
		); err != nil {
			t.Fatalf(
				"não foi possível criar o arquivo %q: %v",
				fileName,
				err,
			)
		}
	}

	catalog := Catalog{
		FormatVersion: CurrentCatalogFormatVersion,
		Datasets: []Dataset{
			{
				Code:        "companies",
				FilePattern: `^Empresas([0-9]+)\.zip$`,
				PartNumberRule: PartNumberRule{
					Source:       PartNumberSourceCaptureGroup,
					CaptureGroup: 1,
				},
			},
			{
				Code:        "countries",
				FilePattern: `^Paises\.zip$`,
				PartNumberRule: PartNumberRule{
					Source: PartNumberSourceFixed,
					Value:  0,
				},
			},
		},
	}

	source := Source{
		Type:     SourceTypeDirectory,
		Location: directory,
	}

	result, err := DiscoverDirectoryPublication(
		catalog,
		source,
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	digest := sha256.Sum256(content)

	expected := []VerifiedFile{
		{
			ClassifiedFile: ClassifiedFile{
				DatasetCode:    "companies",
				PartNumber:     0,
				SourceName:     "Empresas0.zip",
				SourceLocation: filepath.Join(directory, "Empresas0.zip"),
			},
			SizeBytes: uint64(len(content)),
			SHA256:    digest,
		},
		{
			ClassifiedFile: ClassifiedFile{
				DatasetCode:    "companies",
				PartNumber:     1,
				SourceName:     "Empresas1.zip",
				SourceLocation: filepath.Join(directory, "Empresas1.zip"),
			},
			SizeBytes: uint64(len(content)),
			SHA256:    digest,
		},
		{
			ClassifiedFile: ClassifiedFile{
				DatasetCode:    "countries",
				PartNumber:     0,
				SourceName:     "Paises.zip",
				SourceLocation: filepath.Join(directory, "Paises.zip"),
			},
			SizeBytes: uint64(len(content)),
			SHA256:    digest,
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

func TestDiscoverDirectoryPublicationWrapsDiscoveryError(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("não foi possível carregar o catálogo: %v", err)
	}

	source := Source{
		Type:     SourceTypeDirectory,
		Location: filepath.Join(t.TempDir(), "inexistente"),
	}

	result, err := DiscoverDirectoryPublication(
		catalog,
		source,
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != nil {
		t.Errorf(
			"resultado deveria ser nil, mas recebeu: %#v",
			result,
		)
	}

	expectedMessage := "não foi possível descobrir a publicação no diretório"

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf(
			"erro deveria conter %q, mas recebeu: %v",
			expectedMessage,
			err,
		)
	}
}

func TestDiscoverDirectoryPublicationWrapsClassificationError(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "Desconhecido.zip")

	if err := os.WriteFile(
		filePath,
		[]byte("conteúdo de teste"),
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo de teste: %v",
			err,
		)
	}

	catalog := Catalog{
		FormatVersion: CurrentCatalogFormatVersion,
		Datasets: []Dataset{
			{
				Code:        "countries",
				FilePattern: `^Paises\.zip$`,
				PartNumberRule: PartNumberRule{
					Source: PartNumberSourceFixed,
					Value:  0,
				},
			},
		},
	}

	source := Source{
		Type:     SourceTypeDirectory,
		Location: directory,
	}

	result, err := DiscoverDirectoryPublication(
		catalog,
		source,
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != nil {
		t.Errorf(
			"resultado deveria ser nil, mas recebeu: %#v",
			result,
		)
	}

	expectedMessage := "não foi possível classificar a publicação descoberta"

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf(
			"erro deveria conter %q, mas recebeu: %v",
			expectedMessage,
			err,
		)
	}
}

func TestDiscoverDirectoryPublicationWrapsValidationError(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "Empresas0.zip")

	if err := os.WriteFile(
		filePath,
		[]byte("conteúdo de teste"),
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo de teste: %v",
			err,
		)
	}

	catalog := Catalog{
		FormatVersion: CurrentCatalogFormatVersion,
		Datasets: []Dataset{
			{
				Code:        "companies",
				FilePattern: `^Empresas([0-9]+)\.zip$`,
				PartNumberRule: PartNumberRule{
					Source:       PartNumberSourceCaptureGroup,
					CaptureGroup: 1,
				},
			},
			{
				Code:        "countries",
				FilePattern: `^Paises\.zip$`,
				PartNumberRule: PartNumberRule{
					Source: PartNumberSourceFixed,
					Value:  0,
				},
			},
		},
	}

	source := Source{
		Type:     SourceTypeDirectory,
		Location: directory,
	}

	result, err := DiscoverDirectoryPublication(
		catalog,
		source,
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != nil {
		t.Errorf(
			"resultado deveria ser nil, mas recebeu: %#v",
			result,
		)
	}

	expectedMessage := "publicação descoberta no diretório é inválida"

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf(
			"erro deveria conter %q, mas recebeu: %v",
			expectedMessage,
			err,
		)
	}
}
