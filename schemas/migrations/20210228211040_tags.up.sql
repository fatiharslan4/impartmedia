CREATE TABLE IF NOT EXISTS tag
(
    tag_id      INT UNSIGNED AUTO_INCREMENT NOT NULL,
    name        NVARCHAR(64)                NOT NULL,
    long_name   NVARCHAR(255)               NOT NULL,
    description NVARCHAR(1024)              NOT NULL,
    sort_order  INT                         NOT NULL,
    PRIMARY KEY (tag_id),
    unique (name)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

INSERT INTO tag (name, long_name, description, sort_order)
values ('Income', 'Household Income', 'Total pre-tax household income', 10),
       ('Savings', 'Emergency Savings', 'Total household emergency savings', 20),
       ('Education', 'Education Savings', 'Total household education savings', 30),
       ('Retirement', 'Retirement Savings', 'Total household Retirement savings', 40),
       ('Insurance', 'Life Insurance', 'Total Life Insurance Coverage', 50),
       ('Net Worth', 'Net Worth', 'Total Household Net Worth', 60),
       ('Other', 'Other', 'Other', 100)