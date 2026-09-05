package publication

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const CurrentCatalogFormatVersion uint16 = 1

var datasetCodePattern = regexp.MustCompile(
	`^[a-z][a-z0-9_]{0,63}$`,
)

func ValidateCatalog(value Catalog) error {
	var problems []error

	if value.FormatVersion != CurrentCatalogFormatVersion {
		problems = append(
			problems,
			fmt.Errorf(
				"versão do formato do catálogo deveria ser %d, mas recebeu %d",
				CurrentCatalogFormatVersion,
				value.FormatVersion,
			),
		)
	}

	if len(value.Datasets) == 0 {
		problems = append(
			problems,
			errors.New("catálogo de publicações não possui datasets"),
		)
	}

	knownCodes := make(map[string]struct{})
	knownPatterns := make(map[string]struct{})

	for index, dataset := range value.Datasets {
		if !datasetCodePattern.MatchString(dataset.Code) {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d possui código inválido: %q",
					index,
					dataset.Code,
				),
			)
		} else if _, exists := knownCodes[dataset.Code]; exists {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d possui código duplicado: %q",
					index,
					dataset.Code,
				),
			)
		} else {
			knownCodes[dataset.Code] = struct{}{}
		}

		if strings.TrimSpace(dataset.FilePattern) == "" {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d não possui padrão de arquivo",
					index,
				),
			)
			continue
		}

		if _, exists := knownPatterns[dataset.FilePattern]; exists {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d possui padrão de arquivo duplicado: %q",
					index,
					dataset.FilePattern,
				),
			)
		} else {
			knownPatterns[dataset.FilePattern] = struct{}{}
		}

		if !strings.HasPrefix(dataset.FilePattern, "^") ||
			!strings.HasSuffix(dataset.FilePattern, "$") {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d deve possuir padrão de arquivo ancorado",
					index,
				),
			)
		}

		compiledPattern, err := regexp.Compile(dataset.FilePattern)
		if err != nil {
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d possui padrão de arquivo inválido: %w",
					index,
					err,
				),
			)
			continue
		}

		switch dataset.PartNumberRule.Source {
		case PartNumberSourceFixed:
			if dataset.PartNumberRule.CaptureGroup != 0 {
				problems = append(
					problems,
					fmt.Errorf(
						"dataset %d com parte fixa não pode definir grupo de captura",
						index,
					),
				)
			}

		case PartNumberSourceCaptureGroup:
			captureGroup := dataset.PartNumberRule.CaptureGroup

			if captureGroup == 0 ||
				int(captureGroup) > compiledPattern.NumSubexp() {
				problems = append(
					problems,
					fmt.Errorf(
						"dataset %d possui grupo de captura inválido: %d",
						index,
						captureGroup,
					),
				)
			}

			if dataset.PartNumberRule.Value != 0 {
				problems = append(
					problems,
					fmt.Errorf(
						"dataset %d com parte capturada não pode definir valor fixo",
						index,
					),
				)
			}

		default:
			problems = append(
				problems,
				fmt.Errorf(
					"dataset %d possui origem do número da parte inválida: %q",
					index,
					dataset.PartNumberRule.Source,
				),
			)
		}
	}

	return errors.Join(problems...)
}
