-- Seed default mail configuration rows in utm_configuration_parameter so the
-- settings UI has stable keys to bind to before an admin fills them in.
-- Keys mirror the constants in shared/constants/mail_configuration.go
-- (PROP_MAIL_HOST, PROP_MAIL_PORT, PROP_MAIL_PASSWORD, PROP_MAIL_USERNAME,
--  PROP_MAIL_ORGNAME, PROP_MAIL_SMTP_AUTH, PROP_MAIL_FROM, PROP_MAIL_BASE_URL).

INSERT INTO utm_configuration_parameter (key, value, is_secret, label, description, updated_at, updated_by) VALUES
    ('utmstack.mail.host',                          '',     FALSE, 'Mail Host',         'SMTP server hostname',                                NOW(), 'system'),
    ('utmstack.mail.port',                          '587',  FALSE, 'Mail Port',         'SMTP server port (587 STARTTLS, 465 SSL, 25 plain)',  NOW(), 'system'),
    ('utmstack.mail.username',                      '',     FALSE, 'Mail Username',     'SMTP username for authenticated submission',          NOW(), 'system'),
    ('utmstack.mail.password',                      '',     TRUE,  'Mail Password',     'SMTP password (encrypted at rest)',                   NOW(), 'system'),
    ('utmstack.mail.organization',                  '',     FALSE, 'Mail Organization', 'Emitted as the X-Organization header',                NOW(), 'system'),
    ('utmstack.mail.properties.mail.smtp.auth',     'true', FALSE, 'SMTP Auth',         'Whether to authenticate against the SMTP server',     NOW(), 'system'),
    ('utmstack.mail.from',                          '',     FALSE, 'Mail From',         'Sender address; falls back to username when empty',   NOW(), 'system'),
    ('utmstack.mail.baseUrl',                       '',     FALSE, 'Mail Base URL',     'Public panel URL used to render links in emails',     NOW(), 'system')
ON CONFLICT (key) DO NOTHING;
