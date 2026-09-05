package publication

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

//go:embed catalog/datasets.json
var embeddedCatalog []byte

func LoadCatalog() (Catalog, error) {
	return decodeCatalog(embeddedCatalog)
}

func decodeCatalog(data []byte) (Catalog, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var result Catalog

	if err := decoder.Decode(&result); err != nil {
		return Catalog{}, fmt.Errorf(
			"não foi possível decodificar o catálogo de publicações: %w",
			err,
		)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Catalog{}, errors.New(
				"catálogo de publicações possui mais de um documento JSON",
			)
		}

		return Catalog{}, fmt.Errorf(
			"catálogo de publicações possui conteúdo adicional inválido: %w",
			err,
		)
	}

	if err := ValidateCatalog(result); err != nil {
		return Catalog{}, fmt.Errorf(
			"catálogo de publicações inválido: %w",
			err,
		)
	}

	return result, nil
}
