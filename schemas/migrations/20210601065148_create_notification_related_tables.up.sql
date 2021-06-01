-- 
-- Create user device table, , which will hold the data related to the user device
-- 
CREATE TABLE IF NOT EXISTS user_devices (
    token NVARCHAR(250) not null,
    impart_wealth_id CHAR(27) NOT NULL,
    device_id NVARCHAR(250) NOT NULL,
    app_version  NVARCHAR(50) NOT NULL,
    device_name      NVARCHAR(50) NOT NULL,
    device_version   NVARCHAR(50) NOT NULL,
    created_at       DATETIME(3)                    NOT NULL,
    updated_at       DATETIME(3)                    NOT NULL,
    deleted_at       DATETIME(3)                    NULL,
    PRIMARY KEY (token),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;


-- 
-- user_configurations
-- 
-- Which will store user configurations

CREATE TABLE IF NOT EXISTS user_configurations (
    config_id           INT           NOT NULL DEFAULT 0,
    impart_wealth_id    CHAR(27)      NOT NULL,
    notification_status INT           NOT NULL default 0,
    PRIMARY KEY (config_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;

-- 
-- notification_device_mapping
-- 
-- Which will map the device and notification

CREATE TABLE IF NOT EXISTS notification_device_mapping (
    map_id              INT           NOT NULL DEFAULT 0,
    impart_wealth_id    CHAR(27)      NOT NULL,
    user_device_id      NVARCHAR(250)    NOT NULL,
    notify_status       INT           NOT NULL default 0,
    PRIMARY KEY (map_id),
    FOREIGN KEY (impart_wealth_id) REFERENCES user (impart_wealth_id) ON DELETE CASCADE,
    FOREIGN KEY (user_device_id) REFERENCES user_devices (token) ON DELETE CASCADE
) DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  ENGINE = InnoDB
  ROW_FORMAT = DYNAMIC;