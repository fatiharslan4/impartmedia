CREATE INDEX idx_imp_notify_status
ON notification_device_mapping(impart_wealth_id,notify_status);

-- 
CREATE INDEX idx_userdevice_imp_device_token
ON user_devices(impart_wealth_id,device_token);

-- 
CREATE INDEX idx_uconfig_imp_not_status
ON user_configurations(impart_wealth_id,notification_status);