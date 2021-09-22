CREATE TABLE IF NOT EXISTS institutions(
    id                      BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    plaid_institution_id    NVARCHAR(255)       NOT NULL,
    institution_name        NVARCHAR(255)       NOT NULL,
    logo                    NVARCHAR(255)       NOT NULL,
    weburl                  NVARCHAR(255)       NOT NULL,
    PRIMARY KEY (id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


CREATE TABLE IF NOT EXISTS user_institutions(
    institution_id          BIGINT       UNSIGNED NOT NULL,
    impart_wealth_id        CHAR(27)            NOT NULL,
    access_token            NVARCHAR(255)       NOT NULL,
    created_at              DATETIME(3)         NOT NULL,
    PRIMARY KEY (institution_id,impart_wealth_id,access_token),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ,
    FOREIGN KEY (institution_id) REFERENCES institutions (id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;