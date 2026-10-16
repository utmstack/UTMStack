# Compliance Control Library

This folder is the **shared control library**. A framework (PCI DSS, HIPAA, CMMC, ISO 27001,
SOC 2, NIST CSF, NIST 800-171, GDPR, Banking Security Audit) has its own sections and its own
requirements, and each requirement points (`satisfiedBy`) to controls defined **once here**.
The same control serves every framework that needs it: `ac-7` answers PCI DSS 8.3.4,
CMMC AC.L2-3.1.8, NIST 800-171 03.01.08 and ISO A.8.5 alike. There are no per-framework copies.

```
framework  →  section (the framework's own chapter)  →  requirement (its own id)  →  shared controls
pci-dss-4.0   Requirement 10: Log and monitor …        10.2.1.4                     ac-7
```

## Where the controls come from

The controls are the **NIST SP 800-53 Revision 5** catalog (public domain), the catalog almost
every other framework crosswalks to. Source: <https://github.com/usnistgov/oscal-content>.
Base controls use ids like `au-6`; the few control enhancements in the library use the OSCAL
form `ia-2.1`. Controls withdrawn in Rev 5 are not shipped.

| code | Family | code | Family |
|------|--------|------|--------|
| `ac` | Access Control | `mp` | Media Protection |
| `at` | Awareness and Training | `pe` | Physical and Environmental Protection |
| `au` | Audit and Accountability | `pl` | Planning |
| `ca` | Assessment, Authorization, and Monitoring | `pm` | Program Management |
| `cm` | Configuration Management | `ps` | Personnel Security |
| `cp` | Contingency Planning | `pt` | PII Processing and Transparency |
| `ia` | Identification and Authentication | `ra` | Risk Assessment |
| `ir` | Incident Response | `sa` | System and Services Acquisition |
| `ma` | Maintenance | `sc` | System and Communications Protection |
| | | `si` | System and Information Integrity |
| | | `sr` | Supply Chain Risk Management |

The folder name is only organisation; identity is the `id` field.

## Two kinds of control

**Automatic** (`scope: data`) — proven from the event store. Each has checks against logs or
alerts and a `remediation` that says what a failure means and how to fix it.

**Manual** (`scope: governance`) — policies, procedures, training, physical security,
configuration properties logs cannot show (multi-factor authentication, password rules,
deny-by-default). The report lists them as "Out of scope" until a person opens the framework
report, sets the status and writes the justification; the `remediation` says what evidence to
keep. A manual status stays on the report across later evaluations.

A framework requirement takes the **worst** status of its controls, and a manual control ranks
lowest. So a requirement that needs a person's attestation is mapped to manual controls only;
a requirement about recording, monitoring, detecting, reviewing or responding is mapped to the
automatic control that proves it.

## Control file format

```yaml
id: ac-7
family: ac
familyName: Access Control
name: Unsuccessful Logon Attempts
scope: data                   # data | governance
statement: |-                 # the NIST text
  Enforce a limit of …
remediation: |-               # what a failure means and what to do
  …
source: NIST SP 800-53 Rev 5 (public domain)
strategy: ANY                 # ANY: one passing check is enough | ALL: every check must pass
checks:
  - key: win-logon-failures
    name: 'Windows: failed logons and lockouts are recorded'
    dataset: logs             # logs | alerts (default logs)
    dataType: wineventlog     # exact v12 data type; empty means every type
    filters:                  # ANDed; operators from backend/pkg/common_models/filter_type.go
      - field: log.eventCode
        operator: IS_ONE_OF_TERMS
        value: ['4625', '4740']
    rule: MIN_HITS_REQUIRED   # MIN_HITS_REQUIRED (count >= ruleValue) | THRESHOLD_MAX (count <= ruleValue)
    ruleValue: 1
```

How the checks are designed:

- **Every check names one data source (`dataType`).** A check whose source the tenant does not
  send is "not applicable": it neither passes nor fails. A control whose checks are all not
  applicable is "not evaluated", never compliant by default.
- **Evidence checks** (`MIN_HITS_REQUIRED`, ruleValue 1) show something is in place: failed
  logons recorded, backups completing, anti-malware active. One source showing it is enough
  (strategy `ANY`).
- **Limit checks** (`THRESHOLD_MAX`, usually 0) count things that must not happen: an audit log
  cleared, protection switched off, a sign-in without multi-factor, root used directly, telnet or
  FTP allowed. A limit that is exceeded fails the control whatever its strategy, so a clean source
  never hides a violation on another one.
- **Alerts.** `au-6`, `ir-4`, `ir-5`, `si-4` and `ca-7` also read the alert store
  (`dataset: alerts`): alerts raised, reviewed and closed.
- **Field names are what v12 stores**, after the filters in `definitions/filters/`: the
  standard fields (`action`, `actionResult`, `origin.*`, `target.*`) where a filter fills them,
  vendor fields under `log.` otherwise (for example `log.eventCode` for Windows, not
  `log.eventID`). A field outside what the filter writes, or an event the filter drops, never
  matches. A negative operator (`IS_NOT`, `NOT_ENDS_WITH`) also drops rows where the field is
  missing, and `EXIST` on a plain column (`action`, `actionResult`) is always true.
- Checks for sources no server sent yet are built from the filter, the correlation rules and
  the vendor's documentation, and were replayed through the real filters where possible.

## System vs user (templates and customisation)

These shipped files are the **system** layer (read-only, vendor-maintained), refreshed on every
upgrade; do not edit them in place. To customise a control, clone it to the user overlay (the
UI does this on "Edit"). User copies override the system one by `id` and are never overwritten
by upgrades. A control can be disabled with a `.disabled` filename suffix.

## Frameworks

The frameworks live in `../frameworks/`. A requirement lists the controls that satisfy it:

```yaml
- id: '164.312(b)'            # HIPAA Audit controls
  name: Audit controls (Required)
  satisfiedBy: [au-2, au-6, au-12]
```
