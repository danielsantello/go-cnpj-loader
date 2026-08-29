CREATE TABLE publication_files (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    publication_id BIGINT UNSIGNED NOT NULL,
    dataset_code VARCHAR(64) NOT NULL,
    part_number SMALLINT UNSIGNED NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    source_location VARCHAR(2048) NOT NULL,
    size_bytes BIGINT UNSIGNED NULL,
    sha256 BINARY(32) NULL,
    status VARCHAR(16) NOT NULL,
    discovered_at_utc DATETIME(6) NOT NULL,
    verified_at_utc DATETIME(6) NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,

    CONSTRAINT pk_publication_files
        PRIMARY KEY (id),

    CONSTRAINT uq_publication_files_dataset_part
        UNIQUE (
            publication_id,
            dataset_code,
            part_number
        ),

    INDEX idx_publication_files_sha256 (sha256),

    CONSTRAINT fk_publication_files_publication
        FOREIGN KEY (publication_id)
        REFERENCES publications (id),

    CONSTRAINT chk_publication_files_dataset_code
        CHECK (
            dataset_code <> ''
        ),

    CONSTRAINT chk_publication_files_source_name
        CHECK (
            source_name <> ''
        ),

    CONSTRAINT chk_publication_files_source_location
        CHECK (
            source_location <> ''
        ),

    CONSTRAINT chk_publication_files_status
        CHECK (
            status IN (
                'discovered',
                'available',
                'unusable'
            )
        ),

    CONSTRAINT chk_publication_files_lifecycle
        CHECK (
            status = 'discovered'
            OR
            (
                status = 'available'
                AND size_bytes > 0
                AND sha256 IS NOT NULL
                AND verified_at_utc IS NOT NULL
            )
            OR
            (
                status = 'unusable'
                AND verified_at_utc IS NOT NULL
            )
        ),

    CONSTRAINT chk_publication_files_verified_at
        CHECK (
            verified_at_utc IS NULL
            OR verified_at_utc >= discovered_at_utc
        ),

    CONSTRAINT chk_publication_files_status_changed_at
        CHECK (
            status_changed_at_utc >= discovered_at_utc
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
