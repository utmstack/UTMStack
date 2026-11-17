-- DatasetAlerts — id
CREATE TABLE IF NOT EXISTS {{.db}}.{{.alerts}} (
       id String,
       tenantId String,
       dataType String,
       status String,
       parentId String,
       name String,
       severity String,
       assignee String,
       impactScore Int32,
       isIncident Bool,
       tags Array(String),
       "@timestamp" DateTime64(3),
       body Dynamic
           ) ENGINE = MergeTree
PARTITION BY toYYYYMM("@timestamp")
ORDER BY (tenantId, toUnixTimestamp("@timestamp"), id)
