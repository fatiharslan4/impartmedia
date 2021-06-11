-- 
-- Create post video table, , which will hold the data related to admin Post Videos

CREATE TABLE IF NOT EXISTS post_videos (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    PRIMARY KEY (id),
    source NVARCHAR(250) NOT NULL,
    reference_id NVARCHAR(250),
    url NVARCHAR(250) not null,
    post_id  BIGINT UNSIGNED  NOT NULL,
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;