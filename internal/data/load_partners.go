package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadPartnersFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadCustomZIPFile(
		ctx,
		connection,
		schemaName,
		"partners",
		file,
		func(handlerName string) string {
			return fmt.Sprintf(
				`
					LOAD DATA LOCAL INFILE 'Reader::%s'
					INTO TABLE %s.partners
					CHARACTER SET latin1
					FIELDS
						TERMINATED BY ';'
						ENCLOSED BY '"'
						ESCAPED BY ''
					LINES TERMINATED BY '\n'
					(
						@cnpj_root,
						@partner_type_code,
						@partner_name,
						@partner_document,
						@partner_qualification_code,
						@entry_date,
						@country_code,
						@legal_representative_document,
						@legal_representative_name,
						@legal_representative_qualification_code,
						@age_range_code
					)
					SET
						cnpj_root = @cnpj_root,
						partner_type_code = @partner_type_code,
						partner_name = @partner_name,
						partner_document = NULLIF(
							@partner_document,
							''
						),
						partner_qualification_code = CASE
							WHEN @partner_qualification_code = ''
								THEN NULL
							ELSE LPAD(
								@partner_qualification_code,
								2,
								'0'
							)
						END,
						entry_date = CASE
							WHEN @entry_date IN (
								'',
								'0',
								'00000000'
							)
								THEN NULL
							ELSE STR_TO_DATE(
								@entry_date,
								'%%Y%%m%%d'
							)
						END,
						country_code = CASE
							WHEN @country_code = ''
								THEN NULL
							ELSE LPAD(
								@country_code,
								3,
								'0'
							)
						END,
						legal_representative_document = NULLIF(
							@legal_representative_document,
							''
						),
						legal_representative_name = NULLIF(
							@legal_representative_name,
							''
						),
						legal_representative_qualification_code =
							CASE
								WHEN
									@legal_representative_qualification_code = ''
									THEN NULL
								ELSE LPAD(
									@legal_representative_qualification_code,
									2,
									'0'
								)
							END,
						age_range_code = NULLIF(
							@age_range_code,
							''
						)
				`,
				handlerName,
				schemaName,
			)
		},
	)
}
