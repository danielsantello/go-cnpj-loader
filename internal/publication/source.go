package publication

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

type SourceType string

const (
	SourceTypeURL       SourceType = "url"
	SourceTypeDirectory SourceType = "directory"
)

type Source struct {
	Type     SourceType
	Location string
}

func ParseSource(value string) (Source, error) {
	trimmedValue := strings.TrimSpace(value)

	if trimmedValue == "" {
		return Source{}, errors.New("origem da publicação é obrigatória")
	}

	if value != trimmedValue {
		return Source{}, errors.New(
			"origem da publicação não pode possuir espaços externos",
		)
	}

	parsedURL, err := url.Parse(value)
	if err == nil &&
		(parsedURL.Scheme == "http" || parsedURL.Scheme == "https") &&
		parsedURL.Host != "" {
		return Source{
			Type:     SourceTypeURL,
			Location: value,
		}, nil
	}

	if filepath.IsAbs(value) {
		return Source{
			Type:     SourceTypeDirectory,
			Location: value,
		}, nil
	}

	return Source{}, errors.New(
		"origem da publicação deve ser uma URL HTTP(S) válida ou um caminho absoluto",
	)
}
