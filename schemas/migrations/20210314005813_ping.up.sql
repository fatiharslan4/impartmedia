CREATE TABLE IF NOT EXISTS pings
(
    ok bool,
    PRIMARY KEY(ok)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB;

INSERT INTO pings (ok) VALUES (true);