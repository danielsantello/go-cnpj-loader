CREATE TABLE execution_configurations (
    execution_id BIGINT UNSIGNED NOT NULL,
    format_version SMALLINT UNSIGNED NOT NULL,
    configuration JSON NOT NULL,
    created_at_utc DATETIME(6) NOT NULL,

    CONSTRAINT pk_execution_configurations
        PRIMARY KEY (execution_id),

    CONSTRAINT fk_execution_configurations_execution
        FOREIGN KEY (execution_id)
        REFERENCES executions (id)
        ON UPDATE RESTRICT
        ON DELETE RESTRICT,

    CONSTRAINT chk_execution_configurations_format_version
        CHECK (
            format_version > 0
        )
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_0900_ai_ci;
