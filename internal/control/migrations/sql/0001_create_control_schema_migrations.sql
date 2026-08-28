CREATE TABLE control_schema_migrations (
    version INT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    checksum BINARY(32) NOT NULL,
    status VARCHAR(16) NOT NULL,
    loader_version VARCHAR(64) NOT NULL,
    loader_commit VARCHAR(64) NOT NULL,
    started_at_utc DATETIME(6) NOT NULL,
    finished_at_utc DATETIME(6) NULL,
    mysql_error_code INT UNSIGNED NULL,
    sql_state CHAR(5) NULL,
    error_message TEXT NULL,

    CONSTRAINT pk_control_schema_migrations
        PRIMARY KEY (version),

    CONSTRAINT uq_control_schema_migrations_name
        UNIQUE (name),

    CONSTRAINT chk_csm_version
        CHECK (version > 0),

    CONSTRAINT chk_csm_name
        CHECK (name <> ''),

    CONSTRAINT chk_csm_status
        CHECK (
            status IN (
                'applying',
                'applied',
                'failed'
            )
        ),

    CONSTRAINT chk_csm_lifecycle
        CHECK (
            (
                status = 'applying'
                AND finished_at_utc IS NULL
                AND mysql_error_code IS NULL
                AND sql_state IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'applied'
                AND finished_at_utc IS NOT NULL
                AND mysql_error_code IS NULL
                AND sql_state IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'failed'
                AND finished_at_utc IS NOT NULL
                AND error_message IS NOT NULL
            )
        ),

    CONSTRAINT chk_csm_finished_at
        CHECK (
            finished_at_utc IS NULL
            OR finished_at_utc >= started_at_utc
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
