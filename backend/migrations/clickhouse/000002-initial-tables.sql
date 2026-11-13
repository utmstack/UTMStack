-- DatasetLogs — id, no raw
CREATE TABLE IF NOT EXISTS  {{.db}}.{{.logs}} (
               id String, tenantId String, dataType String,
               "@timestamp" DateTime64(3), raw String
           ) ENGINE = MergeTree
PARTITION BY toYYYYMM("@timestamp")
ORDER BY (tenantId, toUnixTimestamp("@timestamp"), id)


