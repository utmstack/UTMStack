-- Remove the seeded alert/incident notification recipient config rows.
DELETE FROM utm_configuration_parameter WHERE conf_param_short IN (
    'utmstack.alerts.notification_to',
    'utmstack.alerts.notification_cc'
);
