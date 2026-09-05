CREATE TABLE execution_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    execution_id BIGINT UNSIGNED NOT NULL,
    execution_step_id BIGINT UNSIGNED NULL,
    severity VARCHAR(16) NOT NULL,
    event_code VARCHAR(64) NOT NULL,
    message TEXT NOT NULL,
    details JSON NULL,
    occurred_at_utc DATETIME(6) NOT NULL,

    CONSTRAINT pk_execution_events
        PRIMARY KEY (id),

    INDEX idx_execution_events_execution_time (
        execution_id,
        occurred_at_utc,
        id
    ),

    INDEX idx_execution_events_step_execution_time (
        execution_step_id,
        execution_id,
        occurred_at_utc,
        id
    ),

    CONSTRAINT fk_execution_events_execution
        FOREIGN KEY (execution_id)
        REFERENCES executions (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT fk_execution_events_execution_step
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

    CONSTRAINT chk_execution_events_severity
        CHECK (
            severity IN (
                'info',
                'warning',
                'error'
            )
        ),

    CONSTRAINT chk_execution_events_event_code
        CHECK (
            event_code <> ''
        ),

    CONSTRAINT chk_execution_events_message
        CHECK (
            message <> ''
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
