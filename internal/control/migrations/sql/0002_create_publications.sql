CREATE TABLE publications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    reference_year SMALLINT UNSIGNED NOT NULL,
    reference_month TINYINT UNSIGNED NOT NULL,
    source_location VARCHAR(2048) NOT NULL,
    content_fingerprint BINARY(32) NOT NULL,
    registered_at_utc DATETIME(6) NOT NULL,

    CONSTRAINT pk_publications
        PRIMARY KEY (id),

    CONSTRAINT uq_publications_reference_content
        UNIQUE (
            reference_year,
            reference_month,
            content_fingerprint
        ),

    CONSTRAINT chk_publications_reference_year
        CHECK (
            reference_year BETWEEN 1000 AND 9999
        ),

    CONSTRAINT chk_publications_reference_month
        CHECK (
            reference_month BETWEEN 1 AND 12
        ),

    CONSTRAINT chk_publications_source_location
        CHECK (
            source_location <> ''
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
