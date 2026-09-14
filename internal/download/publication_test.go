package download

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDownloadPublicationCoordinatesFilesAndReturnsSummary(
	t *testing.T,
) {
	destinationDirectory := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(destinationDirectory, "Cnaes.zip"),
		[]byte("old"),
		0o600,
	); err != nil {
		t.Fatalf(
			"não foi possível criar o arquivo reaproveitável: %v",
			err,
		)
	}

	requests := make([]string, 0)

	client := &http.Client{
		Transport: roundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				requests = append(
					requests,
					request.Method+" "+request.URL.Path,
				)

				switch request.Method {
				case "PROPFIND":
					document := `
						<multistatus xmlns="DAV:">
							<response>
								<href>/public.php/webdav/2026-08/Empresas0.zip</href>
								<propstat>
									<prop>
										<getcontentlength>4</getcontentlength>
										<resourcetype/>
									</prop>
									<status>HTTP/1.1 200 OK</status>
								</propstat>
							</response>
							<response>
								<href>/public.php/webdav/2026-08/Cnaes.zip</href>
								<propstat>
									<prop>
										<getcontentlength>3</getcontentlength>
										<resourcetype/>
									</prop>
									<status>HTTP/1.1 200 OK</status>
								</propstat>
							</response>
						</multistatus>
					`

					return &http.Response{
						StatusCode: http.StatusMultiStatus,
						Body: io.NopCloser(
							strings.NewReader(document),
						),
						Header: make(http.Header),
					}, nil

				case http.MethodGet:
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(
							strings.NewReader("data"),
						),
						Header: make(http.Header),
					}, nil

				default:
					t.Fatalf(
						"método HTTP inesperado: %q",
						request.Method,
					)

					return nil, nil
				}
			},
		),
	}

	progressFiles := make([]string, 0)

	result, err := DownloadPublication(
		t.Context(),
		client,
		2026,
		8,
		destinationDirectory,
		func(current int, total int, file File) {
			progressFiles = append(
				progressFiles,
				file.Name,
			)

			if current != len(progressFiles) {
				t.Errorf(
					"posição atual deveria ser %d, mas recebeu %d",
					len(progressFiles),
					current,
				)
			}

			if total != 2 {
				t.Errorf(
					"total deveria ser 2, mas recebeu %d",
					total,
				)
			}
		},
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expectedResult := PublicationResult{
		FileCount:       2,
		DownloadedCount: 1,
		ReusedCount:     1,
		TotalSizeBytes:  7,
	}

	if result != expectedResult {
		t.Errorf(
			"resultado deveria ser %+v, mas recebeu %+v",
			expectedResult,
			result,
		)
	}

	expectedProgressFiles := []string{
		"Cnaes.zip",
		"Empresas0.zip",
	}

	if !reflect.DeepEqual(
		progressFiles,
		expectedProgressFiles,
	) {
		t.Errorf(
			"progresso deveria ser %#v, mas recebeu %#v",
			expectedProgressFiles,
			progressFiles,
		)
	}

	expectedRequests := []string{
		"PROPFIND /public.php/webdav/2026-08/",
		"GET /public.php/webdav/2026-08/Empresas0.zip",
	}

	if !reflect.DeepEqual(requests, expectedRequests) {
		t.Errorf(
			"requisições deveriam ser %#v, mas recebeu %#v",
			expectedRequests,
			requests,
		)
	}
}
