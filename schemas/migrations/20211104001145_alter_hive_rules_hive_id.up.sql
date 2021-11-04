alter table hive_rules
    add column hive_id  BIGINT UNSIGNED NULL,
    ADD FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE;