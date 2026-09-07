package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadSimpleTaxOptions(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadCustomZIPFile(
		ctx,
		connection,
		schemaName,
		"simple_tax_options",
		file,
		func(handlerName string) string {
			return fmt.Sprintf(
				`
					LOAD DATA LOCAL INFILE 'Reader::%s'
					INTO TABLE %s.simple_tax_options
					CHARACTER SET latin1
					FIELDS
						TERMINATED BY ';'
						ENCLOSED BY '"'
						ESCAPED BY ''
					LINES TERMINATED BY '\n'
					(
						@cnpj_root,
						@simple_option_indicator,
						@simple_option_date,
						@simple_exclusion_date,
						@mei_option_indicator,
						@mei_option_date,
						@mei_exclusion_date
					)
					SET
						cnpj_root = @cnpj_root,
						simple_option_indicator = NULLIF(
							@simple_option_indicator,
							''
						),
						simple_option_date = CASE
							WHEN @simple_option_date IN (
								'',
								'0',
								'00000000'
							)
								THEN NULL
							ELSE STR_TO_DATE(
								@simple_option_date,
								'%%Y%%m%%d'
							)
						END,
						simple_exclusion_date = CASE
							WHEN @simple_exclusion_date IN (
								'',
								'0',
								'00000000'
							)
								THEN NULL
							ELSE STR_TO_DATE(
								@simple_exclusion_date,
								'%%Y%%m%%d'
							)
						END,
						mei_option_indicator = NULLIF(
							@mei_option_indicator,
							''
						),
						mei_option_date = CASE
							WHEN @mei_option_date IN (
								'',
								'0',
								'00000000'
							)
								THEN NULL
							ELSE STR_TO_DATE(
								@mei_option_date,
								'%%Y%%m%%d'
							)
						END,
						mei_exclusion_date = CASE
							WHEN @mei_exclusion_date IN (
								'',
								'0',
								'00000000'
							)
								THEN NULL
							ELSE STR_TO_DATE(
								@mei_exclusion_date,
								'%%Y%%m%%d'
							)
						END
				`,
				handlerName,
				schemaName,
			)
		},
	)
}
