CREATE TABLE IF NOT EXISTS hive
(
    hive_id                BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
    name                   NVARCHAR(255)                  NOT NULL,
    description            NVARCHAR(1024)                 NOT NULL,
    pinned_post_id         BIGINT UNSIGNED                NULL,
    tag_comparisons        JSON                           NULL,
    notification_topic_arn NVARCHAR(255)                  NULL,
    hive_distributions     JSON                           NULL,
    PRIMARY KEY (hive_id),
    INDEX(pinned_post_id)
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

INSERT INTO hive (name, description, tag_comparisons) values ('default', 'the default hive every impart user is a member of',
                                                              '[{"tagId":1,"sortOrder":1,"displayScope":"year","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]},{"tagId":2,"sortOrder":2,"displayScope":"household","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]},{"tagId":3,"sortOrder":5,"displayScope":"child","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]},{"tagId":4,"sortOrder":4,"displayScope":"household","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]},{"tagId":5,"sortOrder":3,"displayScope":"household","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]},{"tagId":7,"sortOrder":6,"displayScope":"household","percentiles":[{"percent":25,"highValue":40000},{"percent":50,"highValue":75000},{"percent":75,"highValue":100000},{"percent":100,"highValue":125000}]}]');