-- 
-- Create post video table, , which will hold the data related to admin Post Videos

CREATE TABLE IF NOT EXISTS post_urls (
    id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    PRIMARY KEY (id),
    title NVARCHAR(250) NOT NULL,
    url NVARCHAR(250),
    imageUrl NVARCHAR(250) not null,
    description NVARCHAR(250) not null,
    post_id  BIGINT UNSIGNED  NOT NULL,
    FOREIGN KEY (post_id) REFERENCES post (post_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;