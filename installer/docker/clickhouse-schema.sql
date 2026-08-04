-- UTMStack ClickHouse schema.
--
-- Partitioning follows one rule across every table:
--
--     PARTITION BY toYYYYMM(`@timestamp`)
--     ORDER BY (tenantId, <dimensions>, `@timestamp`)
--
-- tenantId is never in the partition key and always first in the sorting key.
-- See README.md for why; the short version is that a partition key carrying the
-- tenant multiplies partitions by tenant count and makes an INSERT spanning
-- more than 100 tenants fail outright, while giving nothing back — tenant
-- pruning comes from the sorting key.
--
-- PARTITION BY cannot be altered after the fact. Changing it means rebuilding
-- the table, which with years of data is not an operation anyone wants to run.
--
-- Retention is written here rather than applied from outside, so an install
-- that keeps everything on local disk has nothing to configure. Object storage
-- is opt-in and added later, in place: see EnableTiering in the store driver.

CREATE DATABASE IF NOT EXISTS utmstack;

-- Column names are the wire names of the pipeline's protobuf, not snake_case.
-- ClickHouse drops JSON keys that match no column without reporting anything,
-- so a renamed column here does not fail — it makes every row invisible to
-- every tenant query.

CREATE TABLE IF NOT EXISTS utmstack.logs
(
    `tenantId`         LowCardinality(String),
    `@timestamp`       DateTime64(3, 'UTC'),
    `dataType`         LowCardinality(String),
    `dataSource`       String,
    `id`               String,
    `deviceTime`       DateTime64(3, 'UTC') DEFAULT toDateTime64(0, 3, 'UTC'),
    `tenantName`       LowCardinality(String) DEFAULT '',
    `protocol`         LowCardinality(String) DEFAULT '',
    `connectionStatus` LowCardinality(String) DEFAULT '',
    `statusCode`       String DEFAULT '',
    `action`           LowCardinality(String) DEFAULT '',
    `actionResult`     LowCardinality(String) DEFAULT '',
    `severity`         LowCardinality(String) DEFAULT '',
    `errors`           Array(String) DEFAULT [],

    -- origin and target are nested objects on the wire, not flattened columns.
    -- log is the unpredictable half of a record and is why the table does not
    -- need a migration every time a parser learns a new field.
    `origin`           JSON(max_dynamic_paths = 32),
    `target`           JSON(max_dynamic_paths = 32),
    `compliance`       JSON(max_dynamic_paths = 32),
    `log`              JSON(max_dynamic_paths = 512),

    `raw`              String CODEC(ZSTD(3)),

    INDEX idx_id  id  TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_raw raw TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 4
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(`@timestamp`)
ORDER BY (tenantId, dataType, dataSource, `@timestamp`)
TTL toDateTime(`@timestamp`) + INTERVAL 730 DAY DELETE
SETTINGS ttl_only_drop_parts = 1;

-- Deliberately the same retention as logs. An alert's evidence is the logs it
-- points at, and the drill-down reads them by correlation step, so an alert
-- outliving them is a panel that errors.
--
-- The columns are the union of what the alerts plugin writes at creation and
-- what the backend adds afterwards (assignee, notes, history, incident detail).
-- They are the document's own key names, taken from a marshalled alert rather
-- than from the Alert protobuf: the plugin writes AlertFields, which embeds the
-- protobuf and adds a dozen fields of its own, and the two disagree in ways
-- that matter.
CREATE TABLE IF NOT EXISTS utmstack.alerts
(
    `tenantId`          LowCardinality(String),
    `@timestamp`        DateTime64(3, 'UTC'),
    `id`                String,
    `name`              String,
    `lastUpdate`        DateTime64(3, 'UTC') DEFAULT toDateTime64(0, 3, 'UTC'),
    `tenantName`        LowCardinality(String) DEFAULT '',
    `dataType`          LowCardinality(String) DEFAULT '',
    `dataSource`        String DEFAULT '',
    `category`          LowCardinality(String) DEFAULT '',
    `technique`         LowCardinality(String) DEFAULT '',
    `description`       String DEFAULT '',
    `solution`          String DEFAULT '',

    -- Numeric, not the protobuf's "low"/"medium"/"high". AlertFields declares
    -- its own Severity int, which shadows the embedded string one, so the
    -- stored document carries 1/2/3.
    `severity`          UInt8 DEFAULT 0,
    `severityLabel`     LowCardinality(String) DEFAULT '',

    `status`            Int8 DEFAULT 0,
    `statusLabel`       LowCardinality(String) DEFAULT '',
    `statusObservation` String DEFAULT '',

    -- Empty rather than null: the column is String, so "has no parent" is
    -- an equality test against '', not an existence test.
    `parentId`          String DEFAULT '',

    `impactScore`       Int32 DEFAULT 0,
    `errors`            Array(String) DEFAULT [],

    -- Three pairs of near-duplicates, all present in the document. AlertFields
    -- redeclares fields the embedded protobuf already has under a slightly
    -- different name, so both end up written. The backend reads `reference`
    -- and `deduplicatedBy`; the others are read by nothing. Kept because a
    -- migration that quietly drops a field is the failure this whole exercise
    -- keeps running into — removing them belongs in a separate change to the
    -- plugin, not here.
    `reference`         Array(String) DEFAULT [],
    `references`        Array(String) DEFAULT [],
    `deduplicatedBy`    Array(String) DEFAULT [],
    `deduplicateBy`     Array(String) DEFAULT [],
    `groupedBy`         Array(String) DEFAULT [],
    `groupBy`           Array(String) DEFAULT [],

    `adversary`         JSON(max_dynamic_paths = 64),
    `target`            JSON(max_dynamic_paths = 64),
    `impact`            JSON(max_dynamic_paths = 16),
    `incidentDetail`    JSON(max_dynamic_paths = 16),
    `sourceGroup`       JSON(max_dynamic_paths = 8),
    `lastEvent`         JSON(max_dynamic_paths = 64),

    -- Arrays of objects. A JSON column rejects a top-level array outright
    -- ("JSON object should start with '{'"), so these are Array(JSON).
    `events`            Array(JSON),
    `history`           Array(JSON),

    `tags`              Array(String) DEFAULT [],
    `sourceLabels`      Array(String) DEFAULT [],
    `tagRulesApplied`   Array(Int64) DEFAULT [],
    `logs`              Array(String) DEFAULT [],

    `isIncident`        Bool DEFAULT false,
    `notes`             String DEFAULT '' CODEC(ZSTD(3)),
    `assignee`          String DEFAULT '',
    `assetGroupName`    String DEFAULT '',
    `assetGroupId`      Int64 DEFAULT 0,

    INDEX idx_id     id       TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_parent parentId TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_status status   TYPE set(16) GRANULARITY 4
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(`@timestamp`)
-- name before @timestamp because getPreviousAlertId looks up by tenant and
-- name with no time bound at all; putting the timestamp first would make it
-- scan the tenant's whole history.
ORDER BY (tenantId, name, `@timestamp`)
TTL toDateTime(`@timestamp`) + INTERVAL 730 DAY DELETE
-- non_replicated_deduplication_window is what lets the alerts plugin hand an
-- insert a deduplication token. Two replicas that independently raise the same
-- alert write one row rather than two, and neither has to take a lock to find
-- that out. It bounds how many recent blocks are remembered, so it answers
-- "two writers at once" — the seven-day duplicate check is still a query.
SETTINGS ttl_only_drop_parts = 1, non_replicated_deduplication_window = 1000;

-- What the instance reports for usage, and what the licence is metered
-- against. It carries no dependency on the data it counted and is small, so it
-- outlives the logs and is never moved to cold storage.
CREATE TABLE IF NOT EXISTS utmstack.statistics
(
    `tenantId`   LowCardinality(String),
    `@timestamp` DateTime64(3, 'UTC'),
    `type`       LowCardinality(String),
    `dataType`   LowCardinality(String),
    `dataSource` String,
    `count`      UInt64,
    `bytes`      UInt64,

    -- Identifies the cycle a row was written in, so a retry can tell whether
    -- the batch it is about to re-send already landed. Without it, a write that
    -- committed on the server but timed out on the client is re-sent and the
    -- usage these rows meter is counted twice.
    `batchId`    String DEFAULT ''
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(`@timestamp`)
ORDER BY (tenantId, type, dataSource, dataType, `@timestamp`)
TTL toDateTime(`@timestamp`) + INTERVAL 1095 DAY DELETE
SETTINGS ttl_only_drop_parts = 1;
