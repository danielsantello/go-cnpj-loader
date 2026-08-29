CREATE TABLE publications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    reference_year SMALLINT UNSIGNED NOT NULL,
    reference_month TINYINT UNSIGNED NOT NULL,
    source_type VARCHAR(16) NOT NULL,
    source_location VARCHAR(2048) NOT NULL,
    content_fingerprint BINARY(32) NULL,
    status VARCHAR(16) NOT NULL,
    discovered_at_utc DATETIME(6) NOT NULL,
    verified_at_utc DATETIME(6) NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,

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

    CONSTRAINT chk_publications_source_type
        CHECK (
            source_type IN (
                'url',
                'directory'
            )
        ),

    CONSTRAINT chk_publications_source_location
        CHECK (
            source_location <> ''
        ),

    CONSTRAINT chk_publications_status
        CHECK (
            status IN (
                'discovered',
                'available',
                'superseded',
                'unusable'
            )
        ),

    CONSTRAINT chk_publications_lifecycle
        CHECK (
            status = 'discovered'
            OR
            (
                status = 'available'
                AND content_fingerprint IS NOT NULL
                AND verified_at_utc IS NOT NULL
            )
            OR
            (
                status = 'superseded'
                AND content_fingerprint IS NOT NULL
                AND verified_at_utc IS NOT NULL
            )
            OR
            (
                status = 'unusable'
                AND verified_at_utc IS NOT NULL
            )
        ),

    CONSTRAINT chk_publications_verified_at
        CHECK (
            verified_at_utc IS NULL
            OR verified_at_utc >= discovered_at_utc
        ),

    CONSTRAINT chk_publications_status_changed_at
        CHECK (
            status_changed_at_utc >= discovered_at_utc
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
