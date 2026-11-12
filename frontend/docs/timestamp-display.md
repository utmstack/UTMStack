# Timestamp display

Log Explorer and Alerts use the shared date-format preferences from Settings →
Date & time. `UTC` renders UTC, an IANA zone renders that zone, and `system`
uses the browser's local timezone. Changing this preference does not modify
stored timestamps.

Backend date-time strings without a timezone are interpreted as UTC. Explicit
offsets are preserved; sub-millisecond precision is truncated for JavaScript Date.
Components rendering timestamps subscribe through `useDateFormat`, including
memoized log rows and alert echo timelines.

## Verification

From `frontend`, run:

```sh
npm test -- src/shared/lib/datetime.test.tsx src/features/log-explorer/components/log-table.test.tsx
```

In the UI, switch between UTC and Europe/Istanbul and verify log rows, expanded
log headers, alert list/drawer timestamps, and echo timeline labels. Repeat with
the browser in a different timezone; UTC records must still represent the same
instant. Raw JSON remains unchanged.
