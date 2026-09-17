# Alert grouping and deduplication rollout review

This draft changes runtime alert grouping across technologies. Keep it separate
from technology filter corrections and obtain rollout approval after the checks
below. No customer configuration was changed or deployment performed for this
review.

## Correct query and value paths

The wire `plugins.Alert` carries `events[]`. The indexed `AlertFields` document
adds `lastEvent` from the final event. Therefore a rule's
`lastEvent.origin.user` grouping key must produce an OpenSearch query on that
indexed path while reading its value from the final wire event. Preserve these
two representations; do not replace rule YAML paths with `events.0.*`.

Both `isDuplicate` and `getPreviousAlertId` now share `addAlertGroupingTerms`.
It preserves indexed keys, strips positional array selectors only from the
indexed query path, and rejects null/missing/object/array values. No usable key
means the caller does not execute a name-only search. Other usable keys on the
same rule still participate; rejecting one array field does **not** imply the
whole rule remains unaffected.

Source evidence:

- [Pinned go-sdk v1.1.31 Alert schema](https://github.com/threatwinds/go-sdk/blob/v1.1.31/plugins/plugins.proto)
- [Pinned v11 alert plugin](https://github.com/utmstack/UTMStack/blob/660c796f168670fd6ccb079995c09c46d4f0068d/plugins/alerts/main.go)
- Local regression tests: `plugins/alerts/grouping_test.go`.

## Exact static scope

[`grouping-inventory-v11.json`](grouping-inventory-v11.json) is generated from
**v11 commit `660c796f168670fd6ccb079995c09c46d4f0068d`**, not the working tree.
It records every configured grouping/deduplication path, its rule file/name,
configuration kind and data types. YAML parsing excludes occurrences in prose,
`where`, and historical event queries.

| Measurement | Count |
|---|---:|
| YAML rule files scanned | 635 |
| Files with any groupBy/deduplicateBy setting | 622 |
| All configured field occurrences | 1,177 |
| Files with at least one `lastEvent.*` key | 231 |
| Of those, groupBy files | 227 |
| Of those, deduplicateBy files | 4 |
| `lastEvent.*` occurrences | 309 |
| Distinct `lastEvent.*` paths | 126 |

The 231 files are **potential query changes**, not 231 proven live activations.
They may be disabled on an instance, receive no qualifying events, or use absent
or non-scalar values. Actual values, existing working grouping keys, and deployed
rule revisions determine the result. The scalar guard also applies to other
configured paths, which is why the inventory includes all 622 configured files.
A static inventory cannot establish the full runtime population affected by that
guard.

| Rule directory | Files using `lastEvent.*` |
|---|---:|
| antivirus | 50 |
| cisco | 4 |
| cloud | 91 |
| crowdstrike | 13 |
| fortinet | 2 |
| github | 13 |
| ibm | 3 |
| json | 1 |
| linux | 2 |
| macos | 5 |
| nids | 14 |
| office365 | 19 |
| paloalto | 3 |
| pfsense | 2 |
| sophos | 2 |
| syslog | 3 |
| vmware | 2 |
| windows | 2 |

Reproduce with Python 3 and PyYAML from the repository root:

```sh
python3 plugins/alerts/audits/grouping_inventory.py \
  660c796f168670fd6ccb079995c09c46d4f0068d \
  > /tmp/grouping-inventory-v11.json
cmp plugins/alerts/audits/grouping-inventory-v11.json /tmp/grouping-inventory-v11.json
```

In a partial clone, reading pinned rule blobs can download missing public
repository objects. The script never queries customer instances or modifies git.

## Validation completed and limits

Offline tests serialize real SDK alerts with multiple events, build actual SDK
BoolBuilder term clauses, and check that final-event values appear under indexed
`lastEvent.*` names. They cover `.keyword` input suffixes, positional array member
selection, numeric zero, boolean false, older-event-only fields, missing events,
and non-scalars. Query tests use no index list to avoid network access. They do
not prove deployed index mappings or automatic `.keyword` resolution.

The synthetic filter-normalization harness is separate from these runtime query
construction tests. Passing fixtures do not establish closed EventProcessor raw
extraction, deployed filter/rule parity, history search results, or actual alert
generation. Fixtures without `rules` assertions establish no rule match at all.
Do not describe a green module test run as live detection validation.

## Separate rollout gates — pending

1. Pin the deployed plugin, active rule definitions and index mappings on a
   selected staging system. Compare its active rules with the inventory; include
   site-specific rules and direct keys that may contain arrays/objects.
2. Use the **same ordered sanitized stream and equivalent isolated starting
   alert indices** to compare old and candidate grouping behavior. Avoid comparing
   unrelated calendar windows. Preserve the seven-day deduplication horizon and
   grouping's parent/orphan lookup semantics.
3. Record per-rule candidate alerts, saved alerts, duplicate suppressions, parent
   assignments, group cardinality and key presence/type. Compare resolved final-
   event values and generated indexed query fields. Verify unrelated users, hosts
   and targets do not become a single group.
4. Include multi-event alerts; missing/non-scalar keys; one valid plus one missing
   key; repeated keys; numeric/boolean values; array members; mappings with/without
   `.keyword` subfields. Prioritize populated scalar `lastEvent.*` paths.
5. Investigate every unexplained change in stored alert counts. Approve intended
   per-rule changes and a plugin-image/configuration rollback procedure before
   production rollout. SOC notification and deployment require separate user
   authorization; neither was sent or performed here.

**Pending:** controlled staging replay, actual indexed-query matching, per-rule
before/after alert volume and grouping review, and rollout approval. Customer
access during the maintenance review remains read-only.
