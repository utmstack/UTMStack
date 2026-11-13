CREATE TABLE IF NOT EXISTS  {{.db}}.{{.user_profile}}  (
			   tenantId String,
			   user_id String,
			   profile String,
			   snapshotted_at DateTime64(3),
			   "@timestamp" DateTime64(3) MATERIALIZED snapshotted_at
			) ENGINE = MergeTree
PARTITION BY toYYYYMM("@timestamp")
ORDER BY (tenantId, toUnixTimestamp("@timestamp"))
