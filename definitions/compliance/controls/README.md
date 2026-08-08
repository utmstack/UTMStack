# Compliance Control Library

This folder is the **reusable control library** — the hub of the compliance model.
Each control is a building block ("a check") that compliance **frameworks** reference
(via `satisfiedBy`) and that correlation rules tag (via `compliance:`). A control is
defined **once here** and reused across HIPAA, PCI, CMMC, ISO, etc. — no duplication.

## Where the controls come from

Controls are the official **NIST SP 800-53 Revision 5** catalog — a US government
standard (NIST), **public domain**, and the de-facto backbone of the security-controls
world. It is mandatory for US federal systems (FISMA), the basis of FedRAMP, and the
catalog that almost every other framework crosswalks to. That is why it is our hub: it
is the industry common denominator and is free to use and modify.

Source: <https://github.com/usnistgov/oscal-content> (NIST OSCAL content, public domain).

## Folder layout — the 2-letter codes are the 800-53 control families

The control `id` is `<family>-<number>` (e.g. `au-6`). Each folder is one family:

| code | Family |
|------|--------|
| `ac` | Access Control |
| `at` | Awareness and Training |
| `au` | Audit and Accountability |
| `ca` | Assessment, Authorization, and Monitoring |
| `cm` | Configuration Management |
| `cp` | Contingency Planning |
| `ia` | Identification and Authentication |
| `ir` | Incident Response |
| `ma` | Maintenance |
| `mp` | Media Protection |
| `pe` | Physical and Environmental Protection |
| `pl` | Planning |
| `pm` | Program Management |
| `ps` | Personnel Security |
| `pt` | PII Processing and Transparency |
| `ra` | Risk Assessment |
| `sa` | System and Services Acquisition |
| `sc` | System and Communications Protection |
| `si` | System and Information Integrity |
| `sr` | Supply Chain Risk Management |

The folder name is purely organizational — **identity is the `id` field**, not the path.

## Control file format

```yaml
id: au-6                      # canonical identity (= file/path-independent)
family: au
familyName: Audit and Accountability
name: Audit Record Review, Analysis, and Reporting
scope: data                   # data | governance (default data)
statement: |-                 # what the control requires (NIST text)
  Review and analyze system audit records ...
remediation: |-               # OPTIONAL — what to do when the checks fail
  Enable the ... policy so ...
source: NIST SP 800-53 Rev 5 (public domain)
strategy: ANY                 # ALL | ANY over the checks (default ALL)

checks:                       # OPTIONAL — how the control is proven against the data
  - key: audit-records-reviewed
    name: Audit records are being produced
    dataset: logs             # logs | alerts (default logs)
    dataType: wineventlog     # empty means every type in the dataset
    filters:                  # same filter vocabulary as the log explorer
      - field: log.eventID
        operator: IS_ONE_OF_TERMS
        value: ['4719', '1102']
    rule: MIN_HITS_REQUIRED   # MIN_HITS_REQUIRED (count >= ruleValue) | THRESHOLD_MAX (count <= ruleValue)
    ruleValue: 1
```

A check separates **selecting** from **judging**: the filters say which records count,
the rule says how many of them make a pass. `dataset` + `dataType` double as the
applicability declaration — a check against a data type the tenant does not receive
cannot be failed, only left unevaluated.

Controls that cannot be proven from log data carry `scope: governance` and no checks;
they are excluded from the score rather than counted as failing. A placeholder check
marked `todo: true` means "not defined yet" and is likewise excluded.

## How a control's compliance status is computed (report time, backend)

1. **Coverage** — correlation rules tagged with this control's id (`compliance: [au-6]`).
2. **Activity** — alerts from those rules in the reporting window.
3. **Analysis** — this control's checks over the event store → pass/fail.

The detection/alert pipeline is **not** involved at evaluation time.

## System vs user (templates & customization)

These shipped files are the **system** layer (read-only, vendor-maintained). They are
refreshed from this repo on every upgrade, so do **not** edit them in place.

To customize a control, **clone it to the user overlay** (the UI does this automatically
on "Edit" — copy-on-write). User copies override the system one by `id` and are **never**
overwritten by upgrades. You can also create brand-new controls in the user overlay, and
disable any control with a `.disabled` filename suffix.

## Frameworks

The checklists that reference these controls live in `../frameworks/`. A framework
requirement lists the controls that satisfy it:

```yaml
- id: "164.312(b)"            # HIPAA Audit Controls
  name: Audit Controls
  satisfiedBy: [au-2, au-3, au-6, au-12]
```
