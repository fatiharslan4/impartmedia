CREATE TABLE IF NOT EXISTS hive_rules(
    rule_id                     BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    name                        NVARCHAR(255)       NOT NULL,
    status                      BOOL                NOT NULL,
    max_limit                   INT              NOT NULL,
    no_of_users                 INT              NOT NULL,
    PRIMARY KEY (rule_id),
    UNIQUE (name)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


CREATE TABLE IF NOT EXISTS hive_rules_criteria(
    rule_id                      BIGINT         UNSIGNED     NOT NULL,
    question_id                  INT         UNSIGNED     NOT NULL,
    answer_id                    INT         UNSIGNED     NOT NULL,
    PRIMARY KEY (rule_id,question_id,answer_id),
    INDEX (rule_id, question_id,answer_id),
    FOREIGN KEY (rule_id) REFERENCES hive_rules (rule_id) ON DELETE CASCADE, 
    FOREIGN KEY (question_id) REFERENCES question (question_id),
    FOREIGN KEY (answer_id) REFERENCES answer (answer_id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


CREATE TABLE IF NOT EXISTS hive_rule_map(
    hive_id                      BIGINT         UNSIGNED     NOT NULL,
    rule_id                      BIGINT         UNSIGNED     NOT NULL,
    PRIMARY KEY (hive_id,rule_id),
    FOREIGN KEY (rule_id) REFERENCES hive_rules (rule_id) ON DELETE CASCADE,
    FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;