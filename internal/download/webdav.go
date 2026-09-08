package download

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	webDAVBaseURL  = "https://arquivos.receitafederal.gov.br/public.php/webdav"
	publicShareKey = "YggdBLfdninEJX9"
)

type webDAVMultistatus struct {
	Responses []webDAVResponse `xml:"response"`
}

type webDAVResponse struct {
	Href      string           `xml:"href"`
	Propstats []webDAVPropstat `xml:"propstat"`
}

type webDAVPropstat struct {
	Properties webDAVProperties `xml:"prop"`
	Status     string           `xml:"status"`
}

type webDAVProperties struct {
	ContentLength string             `xml:"getcontentlength"`
	ETag          string             `xml:"getetag"`
	ResourceType  webDAVResourceType `xml:"resourcetype"`
}

type webDAVResourceType struct {
	Collection *struct{} `xml:"collection"`
}

func ListPublicationFiles(
	ctx context.Context,
	client *http.Client,
	referenceYear uint16,
	referenceMonth uint8,
) ([]File, error) {
	if client == nil {
		return nil, errors.New("cliente HTTP não foi informado")
	}

	if referenceYear < 1000 || referenceYear > 9999 {
		return nil, fmt.Errorf(
			"ano de referência é inválido: %d",
			referenceYear,
		)
	}

	if referenceMonth < 1 || referenceMonth > 12 {
		return nil, fmt.Errorf(
			"mês de referência é inválido: %d",
			referenceMonth,
		)
	}

	endpoint := fmt.Sprintf(
		"%s/%04d-%02d/",
		webDAVBaseURL,
		referenceYear,
		referenceMonth,
	)

	request, err := http.NewRequestWithContext(
		ctx,
		"PROPFIND",
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível criar a consulta da publicação: %w",
			err,
		)
	}

	request.Header.Set("Depth", "1")
	request.SetBasicAuth(publicShareKey, "")

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"não foi possível consultar a publicação: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf(
			"consulta da publicação retornou HTTP %d",
			response.StatusCode,
		)
	}

	var document webDAVMultistatus

	if err := xml.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf(
			"não foi possível interpretar a resposta da publicação: %w",
			err,
		)
	}

	files := make([]File, 0)
	knownNames := make(map[string]struct{})

	for _, item := range document.Responses {
		for _, propstat := range item.Propstats {
			if !strings.Contains(propstat.Status, " 200 ") {
				continue
			}

			if propstat.Properties.ResourceType.Collection != nil {
				continue
			}

			decodedHref, err := url.PathUnescape(item.Href)
			if err != nil {
				return nil, fmt.Errorf(
					"endereço remoto inválido %q: %w",
					item.Href,
					err,
				)
			}

			name := path.Base(decodedHref)

			if !strings.EqualFold(path.Ext(name), ".zip") {
				continue
			}

			sizeBytes, err := strconv.ParseUint(
				propstat.Properties.ContentLength,
				10,
				64,
			)
			if err != nil || sizeBytes == 0 {
				return nil, fmt.Errorf(
					"arquivo remoto %q possui tamanho inválido %q",
					name,
					propstat.Properties.ContentLength,
				)
			}

			if _, exists := knownNames[name]; exists {
				return nil, fmt.Errorf(
					"arquivo remoto %q aparece mais de uma vez",
					name,
				)
			}

			knownNames[name] = struct{}{}

			files = append(
				files,
				File{
					Name:      name,
					SizeBytes: sizeBytes,
					ETag: strings.Trim(
						propstat.Properties.ETag,
						`"`,
					),
				},
			)
		}
	}

	if len(files) == 0 {
		return nil, errors.New(
			"publicação remota não possui arquivos ZIP",
		)
	}

	sort.Slice(
		files,
		func(left int, right int) bool {
			return files[left].Name < files[right].Name
		},
	)

	return files, nil
}
