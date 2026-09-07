package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/danielsantello/go-cnpj-loader/internal/publication"
)

func LoadEstablishmentsFile(
	ctx context.Context,
	connection *sql.DB,
	schemaName string,
	file publication.VerifiedFile,
) (uint64, error) {
	return loadCustomZIPFile(
		ctx,
		connection,
		schemaName,
		"establishments",
		file,
		func(handlerName string) string {
			return fmt.Sprintf(
				`
					LOAD DATA LOCAL INFILE 'Reader::%s'
					INTO TABLE %s.establishments
					CHARACTER SET latin1
					FIELDS
						TERMINATED BY ';'
						ENCLOSED BY '"'
						ESCAPED BY ''
					LINES TERMINATED BY '\n'
					(
						@cnpj_root,
						@branch_number,
						@check_digits,
						@head_office_branch_indicator,
						@trade_name,
						@registration_status_code,
						@registration_status_date,
						@registration_status_reason_code,
						@foreign_city_name,
						@country_code,
						@activity_start_date,
						@main_economic_activity_code,
						@secondary_economic_activities,
						@street_type,
						@street_name,
						@street_number,
						@address_complement,
						@neighborhood,
						@postal_code,
						@state_code,
						@municipality_code,
						@phone_area_code_1,
						@phone_number_1,
						@phone_area_code_2,
						@phone_number_2,
						@fax_area_code,
						@fax_number,
						@email_address,
						@special_status,
						@special_status_date
					)
					SET
						cnpj = CONCAT(
							NULLIF(@cnpj_root, ''),
							NULLIF(@branch_number, ''),
							NULLIF(@check_digits, '')
						),
						cnpj_root = NULLIF(
							@cnpj_root,
							''
						),
						branch_number = NULLIF(
							@branch_number,
							''
						),
						check_digits = NULLIF(
							@check_digits,
							''
						),
						head_office_branch_indicator =
							NULLIF(
								@head_office_branch_indicator,
								''
							),
						trade_name = NULLIF(
							@trade_name,
							''
						),
						registration_status_code =
							NULLIF(
								@registration_status_code,
								''
							),
						registration_status_date = CASE
							WHEN @registration_status_date
								IN ('', '0', '00000000')
								THEN NULL
							ELSE STR_TO_DATE(
								@registration_status_date,
								'%%Y%%m%%d'
							)
						END,
						registration_status_reason_code =
							NULLIF(
								@registration_status_reason_code,
								''
							),
						foreign_city_name = NULLIF(
							@foreign_city_name,
							''
						),
						country_code = NULLIF(
							@country_code,
							''
						),
						activity_start_date = CASE
							WHEN @activity_start_date
								IN ('', '0', '00000000')
								THEN NULL
							ELSE STR_TO_DATE(
								@activity_start_date,
								'%%Y%%m%%d'
							)
						END,
						main_economic_activity_code =
							NULLIF(
								@main_economic_activity_code,
								''
							),
						secondary_economic_activities =
							NULLIF(
								@secondary_economic_activities,
								''
							),
						street_type = NULLIF(
							@street_type,
							''
						),
						street_name = NULLIF(
							@street_name,
							''
						),
						street_number = NULLIF(
							@street_number,
							''
						),
						address_complement = NULLIF(
							@address_complement,
							''
						),
						neighborhood = NULLIF(
							@neighborhood,
							''
						),
						postal_code = NULLIF(
							@postal_code,
							''
						),
						state_code = NULLIF(
							@state_code,
							''
						),
						municipality_code = NULLIF(
							@municipality_code,
							''
						),
						phone_area_code_1 = NULLIF(
							@phone_area_code_1,
							''
						),
						phone_number_1 = NULLIF(
							@phone_number_1,
							''
						),
						phone_area_code_2 = NULLIF(
							@phone_area_code_2,
							''
						),
						phone_number_2 = NULLIF(
							@phone_number_2,
							''
						),
						fax_area_code = NULLIF(
							@fax_area_code,
							''
						),
						fax_number = NULLIF(
							@fax_number,
							''
						),
						email_address = NULLIF(
							@email_address,
							''
						),
						special_status = NULLIF(
							@special_status,
							''
						),
						special_status_date = CASE
							WHEN @special_status_date
								IN ('', '0', '00000000')
								THEN NULL
							ELSE STR_TO_DATE(
								@special_status_date,
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
