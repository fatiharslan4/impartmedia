CREATE TABLE IF NOT EXISTS post_tag
(
    tag_id  INT UNSIGNED AUTO_INCREMENT NOT NULL,
    post_id BIGINT UNSIGNED    NOT NULL,
    PRIMARY KEY (post_id, tag_id),
    INDEX (tag_id, post_id),
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag (tag_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;