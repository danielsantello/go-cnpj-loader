CREATE TABLE executions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    version_id BIGINT UNSIGNED NULL,
    operation VARCHAR(32) NOT NULL,
    environment VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL,
    loader_version VARCHAR(64) NOT NULL,
    loader_commit VARCHAR(64) NOT NULL,
    host_name VARCHAR(255) NOT NULL,
    process_id BIGINT UNSIGNED NOT NULL,
    started_at_utc DATETIME(6) NOT NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,
    finished_at_utc DATETIME(6) NULL,
    error_message TEXT NULL,

    CONSTRAINT pk_executions
        PRIMARY KEY (id),

    CONSTRAINT uq_executions_id_version
        UNIQUE (
            id,
            version_id
        ),

    INDEX idx_executions_version_id (version_id),

    INDEX idx_executions_status_started_at (
        status,
        started_at_utc
    ),

    CONSTRAINT fk_executions_version
        FOREIGN KEY (version_id)
        REFERENCES versions (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT chk_executions_operation
        CHECK (
            operation IN (
                'load',
                'delete_version',
                'clean_workspace',
                'migrate_control'
            )
        ),

    CONSTRAINT chk_executions_environment
        CHECK (
            environment IN (
                'development',
                'benchmark',
                'production'
            )
        ),

    CONSTRAINT chk_executions_status
        CHECK (
            status IN (
                'running',
                'succeeded',
                'failed',
                'interrupted',
                'abandoned'
            )
        ),

    CONSTRAINT chk_executions_loader_version
        CHECK (
            loader_version <> ''
        ),

    CONSTRAINT chk_executions_loader_commit
        CHECK (
            loader_commit <> ''
        ),

    CONSTRAINT chk_executions_host_name
        CHECK (
            host_name <> ''
        ),

    CONSTRAINT chk_executions_process_id
        CHECK (
            process_id > 0
        ),

    CONSTRAINT chk_executions_lifecycle
        CHECK (
            (
                status = 'running'
                AND finished_at_utc IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'succeeded'
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'failed'
                AND finished_at_utc IS NOT NULL
                AND error_message IS NOT NULL
            )
            OR
            (
                status = 'interrupted'
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'abandoned'
                AND finished_at_utc IS NOT NULL
                AND error_message IS NOT NULL
            )
        ),

    CONSTRAINT chk_executions_finished_at
        CHECK (
            finished_at_utc IS NULL
            OR finished_at_utc >= started_at_utc
        ),

    CONSTRAINT chk_executions_status_changed_at
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
