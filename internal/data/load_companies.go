package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadCompaniesFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadCustomZIPFile(
		ctx,
		connection,
		schemaName,
		"companies",
		file,
		func(handlerName string) string {
			return fmt.Sprintf(
				`
					LOAD DATA LOCAL INFILE 'Reader::%s'
					INTO TABLE %s.companies
					CHARACTER SET latin1
					FIELDS
						TERMINATED BY ';'
						ENCLOSED BY '"'
						ESCAPED BY ''
					LINES TERMINATED BY '\n'
					(
						@cnpj_root,
						@legal_name,
						@legal_nature_code,
						@responsible_qualification_code,
						@share_capital,
						@company_size_code,
						@responsible_federative_entity
					)
					SET
						cnpj_root = @cnpj_root,
						legal_name = @legal_name,
						legal_nature_code =
							@legal_nature_code,
						responsible_qualification_code =
							@responsible_qualification_code,
						share_capital = NULLIF(
							REPLACE(
								@share_capital,
								',',
								'.'
							),
							''
						),
						company_size_code = NULLIF(
							@company_size_code,
							''
						),
						responsible_federative_entity =
							NULLIF(
								@responsible_federative_entity,
								''
							)
				`,
				handlerName,
				schemaName,
			)
		},
	)
}
