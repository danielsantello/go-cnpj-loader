CREATE TABLE execution_steps (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    execution_id BIGINT UNSIGNED NOT NULL,
    step_code VARCHAR(64) NOT NULL,
    sequence_number SMALLINT UNSIGNED NOT NULL,
    status VARCHAR(16) NOT NULL,
    created_at_utc DATETIME(6) NOT NULL,
    started_at_utc DATETIME(6) NULL,
    status_changed_at_utc DATETIME(6) NOT NULL,
    finished_at_utc DATETIME(6) NULL,
    error_message TEXT NULL,

    CONSTRAINT pk_execution_steps
        PRIMARY KEY (id),

    CONSTRAINT uq_execution_steps_id_execution
        UNIQUE (
            id,
            execution_id
        ),

    CONSTRAINT uq_execution_steps_sequence
        UNIQUE (
            execution_id,
            sequence_number
        ),

    INDEX idx_execution_steps_status_changed_at (
        status,
        status_changed_at_utc
    ),

    CONSTRAINT fk_execution_steps_execution
        FOREIGN KEY (execution_id)
        REFERENCES executions (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT chk_execution_steps_step_code
        CHECK (
            step_code <> ''
        ),

    CONSTRAINT chk_execution_steps_sequence_number
        CHECK (
            sequence_number > 0
        ),

    CONSTRAINT chk_execution_steps_status
        CHECK (
            status IN (
                'pending',
                'running',
                'succeeded',
                'failed',
                'skipped',
                'interrupted'
            )
        ),

    CONSTRAINT chk_execution_steps_lifecycle
        CHECK (
            (
                status = 'pending'
                AND started_at_utc IS NULL
                AND finished_at_utc IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'running'
                AND started_at_utc IS NOT NULL
                AND finished_at_utc IS NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'succeeded'
                AND started_at_utc IS NOT NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'failed'
                AND started_at_utc IS NOT NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NOT NULL
            )
            OR
            (
                status = 'skipped'
                AND started_at_utc IS NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
            OR
            (
                status = 'interrupted'
                AND started_at_utc IS NOT NULL
                AND finished_at_utc IS NOT NULL
                AND error_message IS NULL
            )
        ),

    CONSTRAINT chk_execution_steps_started_at
        CHECK (
            started_at_utc IS NULL
            OR started_at_utc >= created_at_utc
        ),

    CONSTRAINT chk_execution_steps_finished_at
        CHECK (
            finished_at_utc IS NULL
            OR finished_at_utc >= COALESCE(
                started_at_utc,
                created_at_utc
            )
        ),

    CONSTRAINT chk_execution_steps_status_changed_at
        CHECK (
            status_changed_at_utc >= COALESCE(
                finished_at_utc,
                started_at_utc,
                created_at_utc
            )
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
