CREATE TABLE file_loads (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    publication_id BIGINT UNSIGNED NOT NULL,
    version_id BIGINT UNSIGNED NOT NULL,
    publication_file_id BIGINT UNSIGNED NOT NULL,
    execution_id BIGINT UNSIGNED NOT NULL,
    execution_step_id BIGINT UNSIGNED NOT NULL,
    attempt_number SMALLINT UNSIGNED NOT NULL,
    target_table VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL,
    rows_loaded BIGINT UNSIGNED NULL,
    warning_count INT UNSIGNED NULL,
    started_at_utc DATETIME(6) NOT NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,
    finished_at_utc DATETIME(6) NULL,
    error_message TEXT NULL,

    CONSTRAINT pk_file_loads
        PRIMARY KEY (id),

    CONSTRAINT uq_file_loads_attempt
        UNIQUE (
            version_id,
            publication_id,
            publication_file_id,
            attempt_number
        ),

    CONSTRAINT uq_file_loads_execution_step
        UNIQUE (
            execution_step_id,
            execution_id
        ),

    INDEX idx_file_loads_execution_version (
        execution_id,
        version_id
    ),

    INDEX idx_file_loads_publication_file (
        publication_file_id,
        publication_id
    ),

    INDEX idx_file_loads_status_changed_at (
        status,
        status_changed_at_utc
    ),

    CONSTRAINT fk_file_loads_execution_step
        FOREIGN KEY (
            execution_step_id,
            execution_id
        )
        REFERENCES execution_steps (
            id,
            execution_id
        )
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT fk_file_loads_execution_version
        FOREIGN KEY (
            execution_id,
            version_id
        )
        REFERENCES executions (
            id,
            version_id
        )
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT fk_file_loads_version_publication
        FOREIGN KEY (
            version_id,
            publication_id
        )
        REFERENCES versions (
            id,
            publication_id
        )
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT fk_file_loads_publication_file
        FOREIGN KEY (
            publication_file_id,
            publication_id
        )
        REFERENCES publication_files (
            id,
            publication_id
        )
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT chk_file_loads_attempt_number
        CHECK (
            attempt_number > 0
        ),

    CONSTRAINT chk_file_loads_target_table
        CHECK (
            target_table <> ''
        ),

    CONSTRAINT chk_file_loads_status
        CHECK (
            status IN (
                'running',
                'succeeded',
                'failed',
                'interrupted'
            )
        ),

    CONSTRAINT chk_file_loads_lifecycle
        CHECK (
            (
                status = 'running'
                AND rows_loaded IS NULL
                AND warning_count IS NULL
                AND finished_at_utc IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'succeeded'
                AND rows_loaded IS NOT NULL
                AND warning_count = 0
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'failed'
                AND rows_loaded IS NULL
                AND warning_count IS NOT NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NOT NULL
            )
            OR
            (
                status = 'interrupted'
                AND rows_loaded IS NULL
                AND warning_count IS NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
        ),

    CONSTRAINT chk_file_loads_finished_at
        CHECK (
            finished_at_utc IS NULL
            OR finished_at_utc >= started_at_utc
        ),

    CONSTRAINT chk_file_loads_status_changed_at
        CHECK (
            status_changed_at_utc >= COALESCE(
                finished_at_utc,
                started_at_utc
            )
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
