CREATE TABLE IF NOT EXISTS notification_topic
(
    topic_arn CHAR(255)             NOT NULL,
    topic_name CHAR(255) NOT NULL,
    PRIMARY KEY (topic_arn)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

CREATE TABLE IF NOT EXISTS notification_subscriptions
(
    impart_wealth_id  CHAR(27) NOT NULL,
    topic_arn CHAR(255) NOT NULL,
    subscription_arn CHAR(255) NOT NULL,
    PRIMARY KEY (impart_wealth_id, topic_arn),
    INDEX (topic_arn, impart_wealth_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE,
    FOREIGN KEY (topic_arn) REFERENCES notification_topic(topic_arn)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;