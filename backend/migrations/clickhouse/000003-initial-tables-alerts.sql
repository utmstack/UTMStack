-- DatasetAlerts — id, no raw
CREATE TABLE IF NOT EXISTS {{.db}}.{{.alerts}} (
       id String, tenantId String, dataType String,
       "@timestamp" DateTime64(3)
           ) ENGINE = MergeTree
PARTITION BY toYYYYMM("@timestamp")
ORDER BY (tenantId, toUnixTimestamp("@timestamp"), id)
