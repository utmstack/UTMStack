# Related logs

`GET /api/v1/utm-alerts/related-logs?alertId=<id>` returns distinct event IDs
from the requested alert and its echoes, including their `events` and
`lastEvent` references. Echoes are read through the tenant-scoped alert
repository in pages of 100. The request examines at most 1,000 echoes and
returns at most 1,000 distinct event IDs; `truncated` is true if either bound
omits results. `total` is the number of returned IDs, not an uncapped count.

The response time range covers the collected alert/event timestamps with the
existing one-hour lookback and five-minute forward padding. Both RFC3339
timestamps and UTC event-store timestamps without a zone suffix are accepted,
including fractional seconds. Existing rule correlation lookups remain in
place and can extend the lookback.

Regression tests:

```sh
cd backend
go test ./modules/alerts/...
```
