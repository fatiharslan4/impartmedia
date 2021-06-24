DROP TABLE IF EXISTS notification_subscriptions;

-- 
-- Create new Notification subscribtion table with platform endpoint
-- 
CREATE TABLE IF NOT EXISTS notification_subscriptions(
    ns_id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    impart_wealth_id  CHAR(27) NOT NULL,
    topic_arn CHAR(255) NOT NULL,
    subscription_arn CHAR(255) NOT NULL,
    platform_endpoint_arn CHAR(255) NOT NULL,
    PRIMARY KEY (ns_id),
    UNIQUE (platform_endpoint_arn)    
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC; 