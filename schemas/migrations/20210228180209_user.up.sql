CREATE TABLE IF NOT EXISTS `user`
(
    impart_wealth_id        CHAR(27)      NOT NULL,
    authentication_id       NVARCHAR(50)  NOT NULL,
    email                   NVARCHAR(320) NOT NULL,
    screen_name             NVARCHAR(50)  NOT NULL,
    created_ts              DATETIME(3)   NOT NULL,
    updated_ts              DATETIME(3)   NOT NULL,
    device_token            NVARCHAR(255) NOT NULL,
    aws_sns_app_arn         NVARCHAR(255) NOT NULL,
    admin                   BOOL          NOT NULL,

    PRIMARY KEY (impart_wealth_id),
    UNIQUE (authentication_id),
    UNIQUE (email),
    UNIQUE (screen_name)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;