package download

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return function(request)
}

func TestListPublicationFilesReturnsSortedZIPFiles(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(
			func(request *http.Request) (*http.Response, error) {
				if request.Method != "PROPFIND" {
					t.Errorf(
						"método deveria ser PROPFIND, mas recebeu %q",
						request.Method,
					)
				}

				expectedURL := webDAVBaseURL + "/2026-08/"
				if request.URL.String() != expectedURL {
					t.Errorf(
						"URL deveria ser %q, mas recebeu %q",
						expectedURL,
						request.URL.String(),
					)
				}

				if request.Header.Get("Depth") != "1" {
					t.Errorf(
						"cabeçalho Depth deveria ser 1, mas recebeu %q",
						request.Header.Get("Depth"),
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

				document := `
					<multistatus xmlns="DAV:">
						<response>
							<href>/public.php/webdav/2026-08/</href>
							<propstat>
								<prop>
									<resourcetype>
										<collection/>
									</resourcetype>
								</prop>
								<status>HTTP/1.1 200 OK</status>
							</propstat>
						</response>
						<response>
							<href>/public.php/webdav/2026-08/Empresas1.zip</href>
							<propstat>
								<prop>
									<getcontentlength>20</getcontentlength>
									<getetag>&quot;etag-empresas&quot;</getetag>
									<resourcetype/>
								</prop>
								<status>HTTP/1.1 200 OK</status>
							</propstat>
						</response>
						<response>
							<href>/public.php/webdav/2026-08/README.txt</href>
							<propstat>
								<prop>
									<getcontentlength>30</getcontentlength>
									<resourcetype/>
								</prop>
								<status>HTTP/1.1 200 OK</status>
							</propstat>
						</response>
						<response>
							<href>/public.php/webdav/2026-08/Cnaes.zip</href>
							<propstat>
								<prop>
									<getcontentlength>10</getcontentlength>
									<getetag>&quot;etag-cnaes&quot;</getetag>
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
			},
		),
	}

	result, err := ListPublicationFiles(
		t.Context(),
		client,
		2026,
		8,
	)
	if err != nil {
		t.Fatalf("não esperava erro, mas recebeu: %v", err)
	}

	expected := []File{
		{
			Name:      "Cnaes.zip",
			SizeBytes: 10,
			ETag:      "etag-cnaes",
		},
		{
			Name:      "Empresas1.zip",
			SizeBytes: 20,
			ETag:      "etag-empresas",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf(
			"arquivos deveriam ser %#v, mas recebeu %#v",
			expected,
			result,
		)
	}
}

func TestListPublicationFilesRejectsInvalidInputs(t *testing.T) {
	client := &http.Client{}

	tests := []struct {
		name            string
		client          *http.Client
		referenceYear   uint16
		referenceMonth  uint8
		expectedMessage string
	}{
		{
			name:            "cliente ausente",
			client:          nil,
			referenceYear:   2026,
			referenceMonth:  8,
			expectedMessage: "cliente HTTP não foi informado",
		},
		{
			name:            "ano inválido",
			client:          client,
			referenceYear:   999,
			referenceMonth:  8,
			expectedMessage: "ano de referência é inválido",
		},
		{
			name:            "mês inválido",
			client:          client,
			referenceYear:   2026,
			referenceMonth:  13,
			expectedMessage: "mês de referência é inválido",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ListPublicationFiles(
				t.Context(),
				test.client,
				test.referenceYear,
				test.referenceMonth,
			)
			if err == nil {
				t.Fatal("esperava erro, mas recebeu nil")
			}

			if result != nil {
				t.Errorf(
					"resultado deveria ser nil, mas recebeu %#v",
					result,
				)
			}

			if !strings.Contains(
				err.Error(),
				test.expectedMessage,
			) {
				t.Errorf(
					"erro deveria conter %q, mas recebeu: %v",
					test.expectedMessage,
					err,
				)
			}
		})
	}
}

func TestListPublicationFilesRejectsUnexpectedHTTPStatus(
	t *testing.T,
) {
	client := &http.Client{
		Transport: roundTripFunc(
			func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body: io.NopCloser(
						strings.NewReader("indisponível"),
					),
					Header: make(http.Header),
				}, nil
			},
		),
	}

	result, err := ListPublicationFiles(
		t.Context(),
		client,
		2026,
		8,
	)
	if err == nil {
		t.Fatal("esperava erro, mas recebeu nil")
	}

	if result != nil {
		t.Errorf(
			"resultado deveria ser nil, mas recebeu %#v",
			result,
		)
	}

	expectedMessage := "consulta da publicação retornou HTTP 503"

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf(
			"erro deveria conter %q, mas recebeu: %v",
			expectedMessage,
			err,
		)
	}
}
