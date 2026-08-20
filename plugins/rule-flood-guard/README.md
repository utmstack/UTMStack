# UTMStack Rule Flood Guard Plugin

Safety net against alert fatigue: when a correlation rule floods the alerts
list with individual, non-deduplicated alerts from the same data source,
this plugin automatically disables the rule and notifies the user through
the notification bell, suggesting they mark the alerts as false positives
(Alert Tag Rules) or fine-tune the rule (`deduplicateBy`/`groupBy`).

It runs as its own independent plugin, separate from `plugins/alerts`.

## Multitenancy

Everything is scoped per tenant:

- **Counting.** Alerts are grouped by tenant, rule name and data source. The
  threshold is compared against one tenant's own volume, so two tenants that
  are each below it never add up to a flood between them.
- **Disabling.** The rule is disabled only for the tenant that flooded. Every
  other tenant keeps it running.
- **Notifying.** The offending tenant is notified, since the disable applies to
  them and the remediation is theirs to apply. The platform tenant receives a
  copy so the operator keeps instance-wide visibility — unless it is already
  the offending tenant, in which case a single notification is sent.

Alerts that carry no tenant are dropped rather than attributed to a default
tenant.

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
| `threshold` | `50` | How many alerts from the same rule and data source, within a single tenant, trigger the auto-disable. |
| `windowHours` | `24` | Time window used to count alerts. |
| `intervalSeconds` | `300` | How often the guard checks. |

If the file doesn't exist, the plugin creates it with these defaults the first time it starts — it never overwrites a file that's already there.


