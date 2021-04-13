# CREATE TABLE IF NOT EXISTS user_questionnaire
# (
#     impart_wealth_id CHAR(27) NOT NULL ,
#     questionnaire_id INT UNSIGNED NOT NULL,
#     PRIMARY KEY (impart_wealth_id, questionnaire_id),
#     FOREIGN KEY (impart_wealth_id) references user (impart_wealth_id) ON DELETE CASCADE,
#     FOREIGN KEY (questionnaire_id) REFERENCES questionnaire (questionnaire_id) ON DELETE CASCADE
# ) DEFAULT CHARACTER SET utf8mb4
#   COLLATE utf8mb4_unicode_ci
#   ENGINE = InnoDB
#   ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS user_answers
(
    impart_wealth_id CHAR(27) NOT NULL,
    answer_id INT UNSIGNED NOT NULL,
    created_at       DATETIME(3)                    NOT NULL,
    updated_at       DATETIME(3)                    NOT NULL,
    deleted_at       DATETIME(3)                    NULL,
    PRIMARY KEY (impart_wealth_id, answer_id),
    FOREIGN KEY (impart_wealth_id) references user (impart_wealth_id) ON DELETE CASCADE,
    FOREIGN KEY (answer_id) REFERENCES answer (answer_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;