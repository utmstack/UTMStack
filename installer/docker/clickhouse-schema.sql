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

    -- The protobuf's own "low"/"medium"/"high", not a parallel numeric code.
    -- An enum rather than a String so that ordering by severity is by rank and
    -- not alphabetical, which would read high < low < medium. The plugin
    -- normalises before writing: a value outside the enum fails the insert.
    `severity`          Enum8('low' = 1, 'medium' = 2, 'high' = 3) DEFAULT 'low',

    -- The label is the value. Status has no rank to preserve, so it stays a
    -- string; the backend writes its own states here as the alert is worked.
    `status`            LowCardinality(String) DEFAULT '',
    `statusObservation` String DEFAULT '',

    -- Empty rather than null: the column is String, so "has no parent" is
    -- an equality test against '', not an existence test.
    `parentId`          String DEFAULT '',

    `impactScore`       Int32 DEFAULT 0,
    `errors`            Array(String) DEFAULT [],

    -- One name each, the protobuf's. These used to be three pairs of
    -- near-duplicates, written twice because the plugin's struct redeclared
    -- fields the embedded protobuf already had.
    `references`        Array(String) DEFAULT [],
    `deduplicateBy`     Array(String) DEFAULT [],
    `groupBy`           Array(String) DEFAULT [],

    `adversary`         JSON(max_dynamic_paths = 64),
    `target`            JSON(max_dynamic_paths = 64),
    `impact`            JSON(max_dynamic_paths = 16),
    `incidentDetail`    JSON(max_dynamic_paths = 16),

    -- Around 240 rules group or deduplicate by lastEvent.* paths, so lastEvent
    -- is queried by arbitrary parsed log fields and has to carry the whole
    -- event.
    --
    -- events carries whole events too, for a different reason: the alert drawer
    -- renders them as the log documents they are, so an analyst sees what raised
    -- the alert without leaving the page. It is the widest thing an alert holds
    -- — each entry has the log's raw text and its parsed fields — and that is
    -- the cost of the sample being readable rather than a list of ids.
    `lastEvent`         JSON(max_dynamic_paths = 64),
    `events`            Array(JSON),

    -- A JSON column rejects a top-level array outright ("JSON object should
    -- start with '{'"), so this is Array(JSON) too.
    `history`           Array(JSON),

    `tags`              Array(String) DEFAULT [],

    -- Tag-rule ids are UUIDs, like every tenant-scoped key in Postgres. Held as
    -- String rather than UUID: the column records which rules fired, it is never
    -- joined or compared against, and a rule deleted afterwards still has to
    -- read back as whatever the alert recorded.
    `tagRulesApplied`   Array(String) DEFAULT [],
    `logs`              Array(String) DEFAULT [],

    `isIncident`        Bool DEFAULT false,
    `notes`             String DEFAULT '' CODEC(ZSTD(3)),
    `assignee`          String DEFAULT '',

    INDEX idx_id     id       TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_parent parentId TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_status status   TYPE set(16) GRANULARITY 4
)
-- Replacing, not plain MergeTree: an alert is the one record here that changes
-- after it is written. Analysts set its status, notes, assignee and tags, and
-- append to its history, all day. ALTER ... UPDATE would answer that with a
-- mutation per action — asynchronous, rewriting whole parts — so instead the
-- row is rewritten whole and the newest wins, keyed by lastUpdate.
--
-- That makes lastUpdate load-bearing: a row written without it collapses to
-- 1970 and loses to every other version of itself.
ENGINE = ReplacingMergeTree(`lastUpdate`)
PARTITION BY toYYYYMM(`@timestamp`)
-- The sorting key is also the deduplication key, so it has to identify an alert
-- exactly — hence id at the end. Everything before it is the read order worth
-- keeping: name second because getPreviousAlertId looks up by tenant and name
-- with no time bound at all, and putting the timestamp first would make it scan
-- the tenant's whole history.
--
-- Every column here is fixed at creation and never rewritten, which is what
-- lets an update land on top of the row it is updating instead of beside it.
ORDER BY (tenantId, name, `@timestamp`, id)
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


-- ============================================================================
-- Behavioral profiling.
--
-- All three tables below are maintained by ClickHouse itself; the pipeline
-- never writes to them by hand.
--
--   user_daily_activity   incremental MV on logs. One row per (tenant, user,
--                         day): login count, event count, the distinct login
--                         IPs, circular sin/cos sums of login hours and the
--                         last login row. This compressed memory outlives the
--                         730-day TTL of logs, so profiles keep working after
--                         raw data has rotated.
--   user_behavior_profile refreshable MV (REFRESH EVERY 1 HOUR) over the last
--                         90 days of user_daily_activity. One row per user:
--                         mean / stddevSampStable of daily counts, p95,
--                         circular hour mean (sin/cos) and spread, topK(10)
--                         login IPs, last known geo. ReplacingMergeTree: each
--                         refresh replaces the whole table atomically, so a
--                         reader (the user-behavioral plugin) never sees a
--                         half-computed row. No TTL: a user that leaves the
--                         90-day window simply has no row in the next
--                         refresh.
--   user_profile_history  refreshable MV (REFRESH EVERY 1 DAY APPEND) that
--                         snapshots the profile once a day. Long-term memory:
--                         a dormant user that returns is judged against these
--                         rows instead of "no profile at all".
--
-- The login predicate is written twice (the exploration view and the
-- incremental MV) because a ClickHouse MV only fires on inserts into its own
-- FROM table, so it cannot delegate the predicate to the view. Both copies
-- must stay in sync with definitions/filters:
--
--   wineventlog  log.eventCode 4624 ONLY windows-events.yaml:673
--   linux        syslogIdentifier sshd/sshd-session actionResult success, or
--                auditd action USER_LOGIN actionResult success  linux.yaml:927,936,534
--   azure        log.azureKind 'signin'      azure-eventhub.yaml:2947
--   o365         log.RecordType 15 actionResult success (UserLoginFailed is an attack)
--   aws          log.eventSource 'signin.amazonaws.com' + ConsoleLogin aws.yaml:512
--   google       log.protoPayloadServiceName 'login.googleapis.com' gcp.yaml:527, 543
--
-- Events without origin.ip (windows and linux filters do not set it) still
-- feed the volume metric; hour, IP and geo judgments only apply to rows that
-- carry an ip. Geolocation (origin.geolocation.*) is only present on rows
-- from filters that run the geo transform (azure, o365, aws, gcp); every geo
-- judgment degrades silently when it is absent.
-- ============================================================================

CREATE TABLE IF NOT EXISTS utmstack.user_daily_activity
(
    `tenantId`        LowCardinality(String),
    `day`             Date,
    `user`            String,
    `login_count`     UInt64,
    `event_count`     UInt64,
    `login_ips`       Array(String),
    `ip_login_count`  UInt64,
    `sin_sum`         Float64,
    `cos_sum`         Float64,
    `last_login_at`   DateTime64(3, 'UTC'),
    `last_login_ip`   String,
    `last_login_country` String
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(`day`)
ORDER BY (tenantId, user, day)
TTL `day` + INTERVAL 1095 DAY DELETE
SETTINGS ttl_only_drop_parts = 1;

-- Every event of a known user, annotated with its login status. A plain view
-- for ad-hoc exploration and reporting; the incremental MV does NOT read it
-- (a MV only fires on inserts into its own FROM table) and duplicates the
-- predicate — keep the two in sync.
CREATE VIEW IF NOT EXISTS utmstack.user_activity_rows AS
SELECT
    `tenantId`,
    toStartOfDay(`@timestamp`) AS `day`,
    `@timestamp`               AS `ts`,
    toString(`origin`.`user`)  AS `user`,
    toString(`origin`.`ip`)    AS `ip`,
    toString(`origin`.`geolocation`.`countryCode`) AS `geo_country_code`,
    (
         (dataType = 'wineventlog'
          AND toString(`log`.`eventCode`) = '4624')
      OR (dataType = 'linux'
          AND ((toString(`log`.`syslogIdentifier`) IN ('sshd', 'sshd-session')
                AND `actionResult` = 'success')
               OR (toString(`log`.`syslogIdentifier`) = 'auditd'
                   AND toString(`action`) = 'USER_LOGIN'
                   AND `actionResult` = 'success')))
      OR (dataType = 'azure'
          AND lower(toString(`log`.`azureKind`)) = 'signin')
      OR (dataType = 'o365'
          AND toString(`log`.`RecordType`) = '15'
          AND `actionResult` = 'success')
      OR (dataType = 'aws'
          AND toString(`log`.`eventSource`) = 'signin.amazonaws.com'
          AND toString(`log`.`eventName`) = 'ConsoleLogin')
      OR (dataType = 'google'
          AND toString(`log`.`protoPayloadServiceName`) = 'login.googleapis.com')
    ) AS `is_login`
FROM utmstack.logs
WHERE toString(`origin`.`user`) != '';

-- Incremental: fires on every insert block into logs, so
-- user_daily_activity is updated at write time, with zero latency. The login
-- predicate is inlined as a subquery that computes is_login once per row.
CREATE MATERIALIZED VIEW IF NOT EXISTS utmstack.user_daily_activity_mv
TO utmstack.user_daily_activity
AS SELECT
    `tenantId`,
    toStartOfDay(`@timestamp`) AS `day`,
    toString(`origin`.`user`)  AS `user`,
    countIf(`is_login`)        AS `login_count`,
    count()                    AS `event_count`,
    groupUniqArrayIf(toString(`origin`.`ip`),
                     `is_login` AND toString(`origin`.`ip`) != '')   AS `login_ips`,
    countIf(`is_login` AND toString(`origin`.`ip`) != '') AS `ip_login_count`,
    sumIf(sin(toInt32(toHour(`@timestamp`)) * 3.141592653589793 / 12),
          `is_login` AND toString(`origin`.`ip`) != '')      AS `sin_sum`,
    sumIf(cos(toInt32(toHour(`@timestamp`)) * 3.141592653589793 / 12),
          `is_login` AND toString(`origin`.`ip`) != '')      AS `cos_sum`,
    maxIf(`@timestamp`,
          `is_login` AND toString(`origin`.`ip`) != '')      AS `last_login_at`,
    anyIf(toString(`origin`.`ip`),
          `is_login` AND toString(`origin`.`ip`) != '')      AS `last_login_ip`,
    anyIf(toString(`origin`.`geolocation`.`countryCode`),
          `is_login` AND toString(`origin`.`ip`) != ''
              AND toString(`origin`.`geolocation`.`countryCode`) != '')
                                       AS `last_login_country`
FROM (
    SELECT
        `tenantId`,
        `@timestamp`,
        `origin`,
        (
             (dataType = 'wineventlog'
              AND toString(`log`.`eventCode`) = '4624')
          OR (dataType = 'linux'
              AND ((toString(`log`.`syslogIdentifier`) IN ('sshd', 'sshd-session')
                    AND `actionResult` = 'success')
                   OR (toString(`log`.`syslogIdentifier`) = 'auditd'
                       AND toString(`action`) = 'USER_LOGIN'
                       AND `actionResult` = 'success')))
          OR (dataType = 'azure'
              AND lower(toString(`log`.`azureKind`)) = 'signin')
          OR (dataType = 'o365'
              AND toString(`log`.`RecordType`) = '15'
              AND `actionResult` = 'success')
          OR (dataType = 'aws'
              AND toString(`log`.`eventSource`) = 'signin.amazonaws.com'
              AND toString(`log`.`eventName`) = 'ConsoleLogin')
          OR (dataType = 'google'
              AND toString(`log`.`protoPayloadServiceName`) = 'login.googleapis.com')
        ) AS `is_login`
    FROM utmstack.logs
    WHERE toString(`origin`.`user`) != ''
)
GROUP BY `tenantId`, toStartOfDay(`@timestamp`), toString(`origin`.`user`);

CREATE TABLE IF NOT EXISTS utmstack.user_behavior_profile
(
    `tenantId`         LowCardinality(String),
    `user`             String,
    `login_day_mean`   Float64,
    `login_day_std`    Float64,
    `event_day_mean`   Float64,
    `event_day_std`    Float64,
    `event_day_p95`    UInt64,
    `hour_sin`         Float64,
    `hour_cos`         Float64,
    `hour_std`         Float64,
    `top_ips`          Array(String),
    `ip_login_count`   UInt64,
    `main_ip_lat`      Float64,
    `main_ip_lon`      Float64,
    `last_login_at`    DateTime64(3, 'UTC'),
    `last_login_ip`    String,
    `last_login_country` String,
    `min_samples_ok`   Bool,
    `updated_at`       DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(`updated_at`)
PARTITION BY toYYYYMM(`updated_at`)
ORDER BY (tenantId, user)
SETTINGS ttl_only_drop_parts = 1;

-- Hourly: recompute the 90-day statistics for every user and atomically
-- replace the profile. The login hour is circular (00:00 and 23:00 are one
-- hour apart), so the mean hour lives in the atan2 of the mean sin/cos and
-- its spread is the circular standard deviation sqrt(-2 ln R), where R is
-- the mean resultant length. stddevSampStable (not stddevSamp) is used on
-- the daily counts: the plain variant is documented as numerically unstable
-- on wide windows.
--
-- top_ips is the topK(10) login IP of the window. main_ip_lat / main_ip_lon
-- are left as 0: logins only carry a country code, and the impossible-travel
-- judgment only runs when both the last login and the incoming event carry
-- coordinates; the plugin derives the missing coordinate from
-- last_login_country + last_login_ip using the same geolocation lookup it
-- applies to the event.
--
-- Users with no login rows are excluded so the circular mean never divides
-- by zero, and hour_std is guarded against the all-hours-cancel case
-- (R = 0 -> ln(0) = -inf). A user who starts logging in gets a row at the
-- next refresh. Local logins carry no IP, so the circular hour statistics
-- divide by ip_login_count, the subset of logins with an IP; when none have
-- one, hour_std is 0 and the hour judgment stays off for that user.
CREATE MATERIALIZED VIEW IF NOT EXISTS utmstack.user_behavior_profile_mv
REFRESH EVERY 1 HOUR TO utmstack.user_behavior_profile AS
WITH base AS (
    SELECT
        tenantId,
        user,
        sum(login_count)          AS logins_total,
        uniqExact(day)            AS days_active,
        avg(login_count)          AS login_day_mean,
        stddevSampStable(login_count) AS login_day_std,
        avg(event_count)          AS event_day_mean,
        stddevSampStable(event_count) AS event_day_std,
        quantile(0.95)(event_count) AS event_day_p95,
        sum(sin_sum)              AS sin_total,
        sum(cos_sum)              AS cos_total,
        sum(ip_login_count)       AS ip_logins,
        max(last_login_at)        AS last_login_at,
        anyIf(last_login_ip, last_login_ip != '')      AS last_login_ip,
        anyIf(last_login_country, last_login_country != '') AS last_login_country
    FROM utmstack.user_daily_activity
    WHERE day >= toStartOfDay(now()) - INTERVAL 89 DAY
    GROUP BY tenantId, user
    HAVING logins_total > 0
),
top AS (
    SELECT tenantId, user, topK(10)(ip) AS top_ips
    FROM (
        SELECT tenantId, user, arrayJoin(login_ips) AS ip
        FROM utmstack.user_daily_activity
        WHERE day >= toStartOfDay(now()) - INTERVAL 89 DAY
          AND notEmpty(login_ips)
    )
    GROUP BY tenantId, user
)
SELECT
    b.tenantId,
    b.user,
    b.login_day_mean,
    if(b.days_active >= 2, b.login_day_std, 0)   AS login_day_std,
    b.event_day_mean,
    if(b.days_active >= 2, b.event_day_std, 0)   AS event_day_std,
    toUInt64(b.event_day_p95)                     AS event_day_p95,
    if(b.ip_logins > 0, b.sin_total / b.ip_logins, 0) AS hour_sin,
    if(b.ip_logins > 0, b.cos_total / b.ip_logins, 0) AS hour_cos,
    -- Circular std: sqrt(-2 ln R), R the mean resultant length. Guarded
    -- against R = 0 (logins that cancel on the clock) which would be -inf.
    if(b.ip_logins > 0 AND hypot(b.sin_total, b.cos_total) > 0,
       sqrt(-2 * ln(hypot(b.sin_total, b.cos_total) / b.ip_logins)), 0) AS hour_std,
    t.top_ips,
    b.ip_logins AS ip_login_count,
    0 AS main_ip_lat,
    0 AS main_ip_lon,
    b.last_login_at,
    b.last_login_ip,
    b.last_login_country,
    (b.logins_total >= 20 AND b.days_active >= 14) AS min_samples_ok,
    now64(3) AS updated_at
FROM base AS b
LEFT JOIN top AS t ON t.tenantId = b.tenantId AND t.user = b.user;

CREATE TABLE IF NOT EXISTS utmstack.user_profile_history
(
    `tenantId`           LowCardinality(String),
    `user`               String,
    `snapshotted_at`     Date,
    `login_day_mean`     Float64,
    `login_day_std`      Float64,
    `event_day_mean`     Float64,
    `event_day_std`      Float64,
    `event_day_p95`      UInt64,
    `hour_sin`           Float64,
    `hour_cos`           Float64,
    `hour_std`           Float64,
    `top_ips`            Array(String),
    `ip_login_count`     UInt64,
    `main_ip_lat`        Float64,
    `main_ip_lon`        Float64,
    `last_login_at`      DateTime64(3, 'UTC'),
    `last_login_ip`      String,
    `last_login_country` String
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(`snapshotted_at`)
ORDER BY (tenantId, user, snapshotted_at)
TTL `snapshotted_at` + INTERVAL 1095 DAY DELETE
SETTINGS ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW IF NOT EXISTS utmstack.user_profile_history_mv
REFRESH EVERY 1 DAY APPEND TO utmstack.user_profile_history AS
SELECT
    tenantId,
    user,
    toStartOfDay(now()) AS snapshotted_at,
    login_day_mean,
    login_day_std,
    event_day_mean,
    event_day_std,
    toUInt64(event_day_p95) AS event_day_p95,
    hour_sin,
    hour_cos,
    hour_std,
    top_ips,
    ip_login_count,
    main_ip_lat,
    main_ip_lon,
    last_login_at,
    last_login_ip,
    last_login_country
FROM utmstack.user_behavior_profile;
