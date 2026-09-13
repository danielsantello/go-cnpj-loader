CREATE TABLE publication_files (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    publication_id BIGINT UNSIGNED NOT NULL,
    dataset_code VARCHAR(64) NOT NULL,
    part_number SMALLINT UNSIGNED NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    source_location VARCHAR(2048) NOT NULL,
    size_bytes BIGINT UNSIGNED NOT NULL,
    sha256 BINARY(32) NOT NULL,
    registered_at_utc DATETIME(6) NOT NULL,

    CONSTRAINT pk_publication_files
        PRIMARY KEY (id),

    CONSTRAINT uq_publication_files_dataset_part
        UNIQUE (
            publication_id,
            dataset_code,
            part_number
        ),

    CONSTRAINT fk_publication_files_publication
        FOREIGN KEY (publication_id)
        REFERENCES publications (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

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

    CONSTRAINT chk_publication_files_size_bytes
        CHECK (
            size_bytes > 0
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
