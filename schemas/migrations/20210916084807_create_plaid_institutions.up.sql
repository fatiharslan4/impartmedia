CREATE TABLE IF NOT EXISTS institutions(
    id                      BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    plaid_institution_id    NVARCHAR(255)       NOT NULL,
    institution_name        NVARCHAR(255)       NOT NULL,
    logo                    NVARCHAR(255)       NULL,
    weburl                  NVARCHAR(255)       NULL,
    PRIMARY KEY (id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


CREATE TABLE IF NOT EXISTS user_institutions(
    user_institutions_id    BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    institution_id          BIGINT       UNSIGNED NOT NULL,
    impart_wealth_id        CHAR(27)            NOT NULL,
    access_token            NVARCHAR(255)       NOT NULL,
    created_at              DATETIME(3)         NOT NULL,
    PRIMARY KEY (user_institutions_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ,
    FOREIGN KEY (institution_id) REFERENCES institutions (id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;



 CREATE TABLE IF NOT EXISTS user_institution_accounts(
    account_id                BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    user_institutions_id      BIGINT              UNSIGNED NOT NULL,
    plaid_account_id          NVARCHAR(255)       NOT NULL ,
    available                 DECIMAL(65,2)      NOT NULL DEFAULT 0,
    current                   DECIMAL(65,2)      NOT NULL DEFAULT 0,
    iso_currency_code         NVARCHAR(10)        NULL,
    credit_limit                     DECIMAL(65,2)       NULL,
    unofficial_currency_code  NVARCHAR(10)         NULL,
    mask                      NVARCHAR(10)         NULL,
    name                      NVARCHAR(255)        NOT NULL,
    official_name             NVARCHAR(255)        NULL,
    subtype                   NVARCHAR(255)        NULL,
    Type                      NVARCHAR(255)        NULL,
    PRIMARY KEY (account_id),
    FOREIGN KEY (user_institutions_id) REFERENCES user_institutions (user_institutions_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;