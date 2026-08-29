CREATE TABLE versions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    publication_id BIGINT UNSIGNED NOT NULL,
    schema_name VARCHAR(64) NOT NULL,
    environment VARCHAR(16) NOT NULL,
    sequence_number INT UNSIGNED NOT NULL,
    status VARCHAR(16) NOT NULL,
    created_at_utc DATETIME(6) NOT NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,
    ready_at_utc DATETIME(6) NULL,
    failed_at_utc DATETIME(6) NULL,
    deleted_at_utc DATETIME(6) NULL,

    CONSTRAINT pk_versions
        PRIMARY KEY (id),

    CONSTRAINT uq_versions_schema_name
        UNIQUE (schema_name),

    CONSTRAINT uq_versions_publication_environment_sequence
        UNIQUE (
            publication_id,
            environment,
            sequence_number
        ),

    CONSTRAINT fk_versions_publication
        FOREIGN KEY (publication_id)
        REFERENCES publications (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT chk_versions_schema_name
        CHECK (
            schema_name <> ''
        ),

    CONSTRAINT chk_versions_sequence_number
        CHECK (
            sequence_number > 0
        ),

    CONSTRAINT chk_versions_environment
        CHECK (
            environment IN (
                'development',
                'benchmark',
                'production'
            )
        ),

    CONSTRAINT chk_versions_status
        CHECK (
            status IN (
                'pending',
                'loading',
                'ready',
                'failed',
                'deleting',
                'deleted'
            )
        ),

    CONSTRAINT chk_versions_lifecycle
        CHECK (
            (
                status IN (
                    'pending',
                    'loading'
                )
                AND ready_at_utc IS NULL
                AND failed_at_utc IS NULL
                AND deleted_at_utc IS NULL
            )
            OR
            (
                status = 'ready'
                AND ready_at_utc IS NOT NULL
                AND failed_at_utc IS NULL
                AND deleted_at_utc IS NULL
            )
            OR
            (
                status = 'failed'
                AND ready_at_utc IS NULL
                AND failed_at_utc IS NOT NULL
                AND deleted_at_utc IS NULL
            )
            OR
            (
                status = 'deleting'
                AND deleted_at_utc IS NULL
                AND (
                    (
                        ready_at_utc IS NOT NULL
                        AND failed_at_utc IS NULL
                    )
                    OR
                    (
                        ready_at_utc IS NULL
                        AND failed_at_utc IS NOT NULL
                    )
                )
            )
            OR
            (
                status = 'deleted'
                AND deleted_at_utc IS NOT NULL
                AND (
                    (
                        ready_at_utc IS NOT NULL
                        AND failed_at_utc IS NULL
                    )
                    OR
                    (
                        ready_at_utc IS NULL
                        AND failed_at_utc IS NOT NULL
                    )
                )
            )
        ),

    CONSTRAINT chk_versions_event_dates
        CHECK (
            (
                ready_at_utc IS NULL
                OR ready_at_utc >= created_at_utc
            )
            AND
            (
                failed_at_utc IS NULL
                OR failed_at_utc >= created_at_utc
            )
            AND
            (
                deleted_at_utc IS NULL
                OR deleted_at_utc >= created_at_utc
            )
        ),

    CONSTRAINT chk_versions_deleted_at
        CHECK (
            deleted_at_utc IS NULL
            OR deleted_at_utc >= COALESCE(
                ready_at_utc,
                failed_at_utc
            )
        ),

    CONSTRAINT chk_versions_status_changed_at
        CHECK (
            status_changed_at_utc >= COALESCE(
                deleted_at_utc,
                ready_at_utc,
                failed_at_utc,
                created_at_utc
            )
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
