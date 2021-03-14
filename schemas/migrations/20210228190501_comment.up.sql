CREATE TABLE IF NOT EXISTS `comment`
(
    comment_id        BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    post_id           BIGINT UNSIGNED                NOT NULL,
    impart_wealth_id  CHAR(27)                       NOT NULL,
    created_ts        DATETIME(3)                    NOT NULL,
    content           MEDIUMTEXT                     NOT NULL,
    last_reply_ts     DATETIME(3)                    NOT NULL,
    parent_comment_id BIGINT UNSIGNED                NULL,
    up_vote_count     INT                            NOT NULL DEFAULT 0,
    down_vote_count   INT                            NOT NULL DEFAULT 0,
    PRIMARY KEY (comment_id),
    INDEX (post_id, created_ts),
    INDEX (parent_comment_id, last_reply_ts),
    INDEX (impart_wealth_id, created_ts),
    INDEX (post_id, parent_comment_id),
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comment (comment_id) ON DELETE CASCADE,
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = COMPRESSED;

CREATE TABLE IF NOT EXISTS comment_reactions
(
    comment_id       BIGINT UNSIGNED NOT NULL,
    post_id          BIGINT UNSIGNED NOT NULL,
    impart_wealth_id CHAR(27)        NOT NULL,
    upvoted          BOOL            NOT NULL DEFAULT 0,
    downvoted        BOOL            NOT NULL DEFAULT 0,
    updated_ts       DATETIME(3)     NOT NULL,
    PRIMARY KEY (comment_id, impart_wealth_id),
    INDEX (impart_wealth_id, comment_id),
    INDEX (post_id, comment_id, impart_wealth_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comment (comment_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS comment_edits
(
    comment_id       BIGINT UNSIGNED NOT NULL,
    edit_id          BIGINT UNSIGNED NOT NULL,
    impart_wealth_id CHAR(27)        NOT NULL,
    edit_timestamp   DATETIME(3)     NOT NULL,
    deleted          BOOL            NOT NULL DEFAULT 0,
    notes            TEXT            NULL,
    PRIMARY KEY (comment_id, edit_id),
    FOREIGN KEY (comment_id) REFERENCES comment (comment_id) ON DELETE CASCADE,
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;