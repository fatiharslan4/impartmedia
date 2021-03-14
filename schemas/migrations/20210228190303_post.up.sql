CREATE TABLE IF NOT EXISTS post
(
    post_id          BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    hive_id          BIGINT UNSIGNED                NOT NULL,
    impart_wealth_id CHAR(27)                       NOT NULL,
    pinned           BOOL                           NOT NULL,
    created_ts       DATETIME(3)                    NOT NULL,
    subject          NVARCHAR(256)                  NOT NULL,
    content          MEDIUMTEXT                     NOT NULL,
    last_comment_ts  DATETIME(3)                    NOT NULL,
    comment_count    INT                            NOT NULL DEFAULT 0,
    up_vote_count    INT                            NOT NULL DEFAULT 0,
    down_vote_count  INT                            NOT NULL DEFAULT 0,
    PRIMARY KEY (post_id),
    INDEX (hive_id, last_comment_ts DESC),
    INDEX (hive_id, created_ts DESC),
    INDEX (impart_wealth_id, created_ts),
    FOREIGN KEY (hive_id) REFERENCES hive (hive_id) ON DELETE CASCADE,
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE

) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = COMPRESSED;

CREATE TABLE IF NOT EXISTS post_reactions
(
    post_id          BIGINT UNSIGNED NOT NULL,
    impart_wealth_id CHAR(27)        NOT NULL,
    upvoted          BOOL            NOT NULL DEFAULT 0,
    downvoted        BOOL            NOT NULL DEFAULT 0,
    updated_ts       DATETIME(3)     NOT NULL,
    PRIMARY KEY (post_id, impart_wealth_id),
    INDEX (impart_wealth_id, post_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS post_edits
(
    post_id          BIGINT UNSIGNED NOT NULL,
    edit_id          INT UNSIGNED    NOT NULL,
    impart_wealth_id CHAR(27)        NOT NULL,
    edit_timestamp   DATETIME(3)     NOT NULL,
    deleted          BOOL            NOT NULL DEFAULT 0,
    notes            TEXT            NULL,
    PRIMARY KEY (post_id, edit_id),
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE,
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE

) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;