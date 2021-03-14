CREATE TABLE IF NOT EXISTS hive_members
(
    member_hive_id          BIGINT UNSIGNED NOT NULL,
    member_impart_wealth_id CHAR(27)        NOT NULL,
    PRIMARY KEY (member_hive_id, member_impart_wealth_id),
    INDEX (member_impart_wealth_id, member_hive_id),
    FOREIGN KEY (member_hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE,
    FOREIGN KEY (member_impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS hive_admins
(
    admin_hive_id          BIGINT UNSIGNED NOT NULL,
    admin_impart_wealth_id CHAR(27)        NOT NULL,
    PRIMARY KEY (admin_hive_id, admin_impart_wealth_id),
    INDEX (admin_impart_wealth_id, admin_hive_id),
    FOREIGN KEY (admin_hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE,
    FOREIGN KEY (admin_impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;