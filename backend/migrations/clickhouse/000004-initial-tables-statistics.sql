-- DatasetStats
CREATE TABLE IF NOT EXISTS {{.db}}.{{.stats}} (
               tenantId String, type String, topic String, dataSource String, dataType String, batchId String,
               count UInt64, bytes UInt64,
               "@timestamp" DateTime64(3)
           ) ENGINE = MergeTree
PARTITION BY toYYYYMM("@timestamp")
ORDER BY (tenantId, topic, dataSource, dataType, toUnixTimestamp64Milli("@timestamp"))
