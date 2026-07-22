# UTMStack Rule Flood Guard Plugin

Safety net against alert fatigue: when a correlation rule floods the alerts
list with individual, non-deduplicated alerts from the same data source,
this plugin automatically disables the rule and notifies the user through
the notification bell, suggesting they mark the alerts as false positives
(Alert Tag Rules) or fine-tune the rule (`deduplicateBy`/`groupBy`).

The notification message names the rule, the data source, and the alert
threshold that was crossed. The backend allows up to 500 characters per
notification, which comfortably fits the fixed template wording plus
realistic rule and data source names, so the message is never truncated.

## Configuration

Settings live in `system_plugins_rule-flood-guard.yaml`, in the same pipeline
config folder as the other plugins' config files. It reloads automatically
when the file changes — no restart needed.

```yaml
plugins:
  rule-flood-guard:
    enabled: true
    threshold: 50
    windowHours: 24
    intervalSeconds: 300
```

| Field | Default | Meaning |
|---|---|---|
| `enabled` | `true` | Turns the guard on or off. |
| `threshold` | `50` | How many alerts from the same rule and data source trigger the auto-disable. |
| `windowHours` | `24` | Time window used to count alerts. |
| `intervalSeconds` | `300` | How often the guard checks. |

If the file doesn't exist, the guard just runs with these defaults.
