# ClickHouse storage

`schema.sql` is the source of truth for the three tables. Apply it against a
fresh instance:

```sh
clickhouse-client --password "$CLICKHOUSE_PASSWORD" --multiquery < schema.sql
```

## The partitioning standard

Every table follows the same rule:

```sql
PARTITION BY toYYYYMM(`@timestamp`)
ORDER BY (tenantId, <dimensions>, `@timestamp`)
```

**`tenantId` is never in the partition key and always first in the sorting
key.** This is not a style preference; putting the tenant in the partition key
breaks two things.

It multiplies partitions by tenant count. At 100 tenants, a table partitioned by
`(tenantId, day)` with two years of retention reaches 73,000 partitions.
ClickHouse recommends staying under 1,000–10,000 *per table*, because a merge
only ever happens within a partition — beyond that, startup, inserts and queries
all degrade.

It also makes ingest fail outright. `max_partitions_per_insert_block` is 100, so
a batch spanning more than 100 tenants is rejected:

```
DB::Exception: Too many partitions for single INSERT block (more than 100).
```

That is exactly what the events plugin does — one batch of a thousand rows drawn
from every tenant.

Nothing is given up in exchange, because the partition key was never what made
tenant queries fast. Measured on two million rows across two hundred tenants,
partitioned by month: a single-tenant query reads 32,768 rows, 1.6% of the
table. The sparse primary index does the pruning, which is what ClickHouse says
itself in the error above — *"partitioning is not intended to speed up SELECT
queries (ORDER BY key is sufficient to make range queries fast)"*.

What it does cost:

- **Exporting a tenant** is a query rather than a file copy. `FREEZE PARTITION`
  hardlinks a whole partition; without the tenant in the key there is no
  per-tenant partition to freeze. Moving a tenant between instances means
  selecting its rows out.
- **TTL is coarser.** A part is dropped once every row in it has expired, so
  monthly partitions can hold data up to a month past its date. Against 730- and
  1095-day retentions that is under 4%.

Deleting a tenant is unaffected: the driver already does it with
`ALTER TABLE ... DELETE WHERE`, not `DROP PARTITION`.

`PARTITION BY` cannot be altered. Changing it means rebuilding the table, so it
is worth getting right while the tables are empty.

## Retention per table

| Table | Deleted after | Why |
|---|---|---|
| `logs` | 730 days | The volume |
| `alerts` | 730 days | **Matches `logs` on purpose** |
| `statistics` | 1095 days | Small, self-contained, and the billing record |

`alerts` is aligned with `logs` because an alert's evidence *is* the logs: the
drill-down reads them by correlation step. An alert that outlives them is a
panel that errors. Raising alert retention on its own — for a compliance
requirement to show what was detected long after the raw data is gone — needs
the panel to degrade gracefully first.

`statistics` goes the other way. It is what the instance reports to the
Customers Manager for usage and what shows growth over time, it carries no
dependency on the data it counted, and it is small enough that cold storage
would cost more in latency than it saves.

## Without object storage — the default

Nothing to deploy. `schema.sql` carries a plain TTL:

```sql
TTL toDateTime(`@timestamp`) + INTERVAL 730 DAY DELETE
SETTINGS ttl_only_drop_parts = 1
```

`ttl_only_drop_parts` is what makes expiry cheap: ClickHouse waits until every
row in a part has expired and deletes the directory, rather than rewriting parts
to remove rows from them.

## With object storage

Deploy `storage-tiering.xml` into `/etc/clickhouse-server/config.d/`, then the
tables carry two clauses instead of one:

```sql
TTL toDateTime(`@timestamp`) + INTERVAL 90 DAY  TO VOLUME 'cold',
    toDateTime(`@timestamp`) + INTERVAL 730 DAY DELETE
SETTINGS storage_policy = 'hot_cold', ttl_only_drop_parts = 1
```

Data on the cold volume is **still part of the table**. A query reaching into it
is slower; it is not a restore. There is no thaw step and no separate archive to
decide about in advance.

## Adding object storage later

Supported, in place, without rebuilding tables — provided the volume naming rule
in `storage-tiering.xml` was followed. The driver does both steps in the right
order:

```go
err := driver.EnableTiering(ctx, dataset, "hot_cold", store.Retention{
    Keep:      730 * 24 * time.Hour,
    ColdAfter: 90 * 24 * time.Hour,
})
```

They cannot be one statement: the policy has to be adopted before a TTL may name
one of its volumes.
