-- 000003_alert_notification_recipients.up.sql
--
-- Seed the two config rows that hold the global alert/incident email recipients.
-- Both are optional, comma-separated lists. When both are empty the incident
-- mailer falls back to every activated user (see modules/incidents mailer).
--
-- The appconfig usecase only UPDATEs pre-seeded rows, so these must exist for
-- GET/PUT /config/<key> to work.
INSERT INTO utm_configuration_parameter
    (conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
FROM (VALUES
    ('utmstack.alerts.notification_to', 'Alerts notification To',  'Comma-separated addresses that receive alert/incident notifications. Empty falls back to every activated user.', '', false, 'text', NULL),
    ('utmstack.alerts.notification_cc', 'Alerts notification Cc',  'Comma-separated addresses copied on alert/incident notifications.',                                          '', false, 'text', NULL)
) AS v(conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
WHERE NOT EXISTS (
    SELECT 1 FROM utm_configuration_parameter c WHERE c.conf_param_short = v.conf_param_short
);
