ALTER TABLE utm_source_scan
ADD COLUMN next_execution_date TIMESTAMP;

UPDATE utm_source_scan
SET next_execution_date = last_execution_date;

ALTER TABLE utm_source_scan
DROP COLUMN last_execution_date;

