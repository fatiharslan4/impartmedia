CREATE TABLE IF NOT EXISTS profile
(
    impart_wealth_id CHAR(27)    NOT NULL,
    updated_ts       DATETIME(3) NOT NULL,
    attributes       JSON        NOT NULL,
    survey_responses JSON        NOT NULL,
    PRIMARY KEY (impart_wealth_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;
