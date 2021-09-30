ALTER TABLE user_institutions DROP FOREIGN KEY user_institutions_ibfk_1 ; 
ALTER TABLE user_institutions DROP FOREIGN KEY user_institutions_ibfk_2 ; 
ALTER TABLE user_institutions DROP PRIMARY KEY ;
ALTER TABLE user_institutions ADD user_institution_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE user_institutions ADD FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ;
ALTER TABLE user_institutions ADD FOREIGN KEY (institution_id) REFERENCES institutions (id) ;

CREATE TABLE IF NOT EXISTS user_plaid_accounts_log(
    id                          BIGINT              UNSIGNED AUTO_INCREMENT    NOT NULL,
    user_institution_id         BIGINT         UNSIGNED     NOT NULL,
    account_id                  NVARCHAR(255)       NOT NULL,
    mask                        NVARCHAR(10)        NOT NULL,
    name                        NVARCHAR(255)       NOT NULL,
    official_name               NVARCHAR(255)       NOT NULL,
    subtype                     NVARCHAR(100)       NOT NULL,
    type                        NVARCHAR(100)       NOT NULL,
    iso_currency_code           NVARCHAR(100)       NOT NULL,
    unofficial_currency_code    NVARCHAR(100)       NOT NULL,
    available                   DOUBLE              NOT NULL,
    current                     DOUBLE              NOT NULL,
    credit_limit                DOUBLE              NOT NULL,
    created_at                  DATETIME(3)         NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_institution_id) REFERENCES user_institutions (user_institution_id) 
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

  CREATE UNIQUE INDEX index_access_token ON user_institutions(institution_id,impart_wealth_id,access_token);