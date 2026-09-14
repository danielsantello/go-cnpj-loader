package download

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileWritesAndPublishesFile(t *testing.T) {
	content := []byte("zip-test")
	file := File{
		Name:      "Cnaes.zip",
		SizeBytes: uint64(len(content)),
		ETag:      "etag-cnaes",
	}

	client := &http.Client{
		Transport: roundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				if request.Method != http.MethodGet {
					t.Errorf(
						"método deveria ser GET, mas recebeu %q",
						request.Method,
					)
				}

				expectedURL := webDAVBaseURL +
					"/2026-08/Cnaes.zip"

				if request.URL.String() != expectedURL {
					t.Errorf(
						"URL deveria ser %q, mas recebeu %q",
						expectedURL,
						request.URL.String(),
					)
				}

				user, password, ok := request.BasicAuth()
				if !ok ||
					user != publicShareKey ||
					password != "" {
					t.Errorf(
						"autenticação básica inesperada: usuário=%q senha=%q ok=%t",
						user,
						password,
						ok,
					)
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(
						bytes.NewReader(content),
					),
					Header: make(http.Header),
				}, nil
			},
		),
	}

	destinationDirectory := t.TempDir()

	downloaded, err := DownloadFile(
		t.Context(),
		client,
		2026,
		8,
		file,
		destinationDirectory,
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if !downloaded {
		t.Error("arquivo deveria ter sido baixado")
	}

	destinationPath := filepath.Join(
		destinationDirectory,
		file.Name,
	)

	result, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf(
			"não foi possível ler o arquivo baixado: %v",
			err,
		)
	}

	if !bytes.Equal(result, content) {
		t.Errorf(
			"conteúdo deveria ser %q, mas recebeu %q",
			content,
			result,
		)
	}

	partialPath := destinationPath + ".part"

	if _, err := os.Stat(partialPath); !os.IsNotExist(err) {
		t.Errorf(
			"arquivo temporário %q não deveria existir",
			partialPath,
		)
	}
}

func TestDownloadFileReusesExistingFile(t *testing.T) {
	content := []byte("zip-test")
	file := File{
		Name:      "Cnaes.zip",
		SizeBytes: uint64(len(content)),
	}

	destinationDirectory := t.TempDir()
	destinationPath := filepath.Join(
		destinationDirectory,
		file.Name,
	)

	if err := os.WriteFile(
		destinationPath,
		content,
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo existente: %v",
			err,
		)
	}

	client := &http.Client{
		Transport: roundTripFunc(
			func(*http.Request) (*http.Response, error) {
				t.Fatal(
					"não deveria realizar requisição HTTP",
				)

				return nil, nil
			},
		),
	}

	downloaded, err := DownloadFile(
		t.Context(),
		client,
		2026,
		8,
		file,
		destinationDirectory,
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	if downloaded {
		t.Error("arquivo existente deveria ter sido reaproveitado")
	}

	result, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf(
			"não foi possível ler o arquivo reaproveitado: %v",
			err,
		)
	}

	if !bytes.Equal(result, content) {
		t.Errorf(
			"conteúdo deveria permanecer %q, mas recebeu %q",
			content,
			result,
		)
	}
}

func TestDownloadFileRemovesIncompletePartialFile(
	t *testing.T,
) {
	content := []byte("incompleto")
	file := File{
		Name:      "Cnaes.zip",
		SizeBytes: uint64(len(content) + 1),
	}

	client := &http.Client{
		Transport: roundTripFunc(
			func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(
						bytes.NewReader(content),
					),
					Header: make(http.Header),
				}, nil
			},
		),
	}

	destinationDirectory := t.TempDir()

	downloaded, err := DownloadFile(
		t.Context(),
		client,
		2026,
		8,
		file,
		destinationDirectory,
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if downloaded {
		t.Error("arquivo incompleto não deveria ser baixado")
	}

	destinationPath := filepath.Join(
		destinationDirectory,
		file.Name,
	)
	partialPath := destinationPath + ".part"

	if _, err := os.Stat(destinationPath); !os.IsNotExist(err) {
		t.Errorf(
			"arquivo final %q não deveria existir",
			destinationPath,
		)
	}

	if _, err := os.Stat(partialPath); !os.IsNotExist(err) {
		t.Errorf(
			"arquivo temporário %q não deveria existir",
			partialPath,
		)
	}
}
