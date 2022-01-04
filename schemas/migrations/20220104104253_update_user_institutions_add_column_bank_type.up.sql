CREATE TABLE IF NOT EXISTS bank_types(
    bank_type_id               BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    bank_type                  NVARCHAR(100) NOT NULL,
    PRIMARY KEY (bank_type_id),
    INDEX (bank_type_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

insert into bank_types (bank_type) values('normal');
insert into bank_types (bank_type) values('investment');



alter table user_institutions
    add column bank_type  BIGINT UNSIGNED NOT NULL DEFAULT 1, 
    ADD FOREIGN KEY (bank_type) REFERENCES bank_types (bank_type_id) ON DELETE CASCADE;