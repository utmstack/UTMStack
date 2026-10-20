-- EDR permissions
INSERT INTO permissions (name, description) VALUES
    ('edr.read',            'Read EDR data'),
    ('edr.write',           'Write or change EDR configurations')
ON CONFLICT (name) DO NOTHING;
