# Reviewing retained quarantine recovery cases

`utmstack_edr quarantine inspect <id>` captures the admission, restore and purge
journals for one quarantine identifier (1–128 ASCII letters, digits, underscores
or hyphens; new operations use UUIDs). It opens existing private storage
without running recovery, opening SQLite, creating a log or lock, changing
permissions, or creating a missing store. Run it with the service's authorized
administrative identity. The command emits one JSON report on standard output;
diagnostics go to standard error and an incomplete inspection exits nonzero.

The report records the root, operation-directory and manifest file identities,
the exact manifest bytes as base64, and their SHA-256. `captured` means those
bytes were read within the one MiB limit. It does **not** mean the manifest is
valid or that its claims are proven. Malformed bytes are preserved losslessly.
`absent` means that operation directory was not found during its observation;
an existing operation without a readable regular manifest is `unreadable`.
Oversized files, reparse/symlink entries and unsafe private directories are
refused. This command does not read original files, destinations or backups.
For an operator hold, the purge location contains a regular decision file;
inspection reports `entry_kind: file` and captures up to two MiB plus 64 KiB.
Its `directory_id` is the containing purge directory identity and its
`manifest_id` is the decision file identity. Ordinary journals retain their
one MiB limit. Capturing a decision file does not validate its contents.

## Capture before attempting recovery

1. Obtain the UUID from the existing recovery event or error. Arrange an approved
   maintenance window and stop every service/management process that can mutate
   the store. Verify the actual service process has exited. Inspection takes no
   item lock; its three journal reads are separate observations and are not an
   atomic snapshot of a running service.
2. Preserve a protected forensic copy of the database with its WAL/SHM files,
   the quarantine store, recorded raw capture directories, and relevant parent
   directory evidence. Use the organization's normal backup procedure and retain
   file identities and Windows owner/DACL information where possible. Preserve
   originals. The JSON report is journal evidence, not a complete backup.
3. Save `quarantine inspect <id>` standard output and standard error separately
   into the protected case record and record its exit status. Keep an incomplete
   report too. Do not run `quarantine list`, `recover`, `restore` or `purge` first:
   those existing commands open a store that attempts recovery.
4. Compare the saved manifest identities with independently collected directory
   and file identities and any durable database completion receipt. A pathname,
   matching digest, timestamp, or plausible file content alone does not prove
   that an object belongs to this operation. Preserve the exact journal bytes;
   do not edit them to make validation pass.
5. If the recorded identities and approved operation are intact, an operator can
   explicitly run `quarantine recover` and retain its results. Repeat inspection
   after a failure. Recovery is allowed to complete an already approved purge
   and delete its proven artifacts; it is not a read-only diagnostic command.

For example, an administrator can decode a captured journal for review in
PowerShell after saving the unmodified inspection JSON:

```powershell
$case = Get-Content -Raw .\inspection.json | ConvertFrom-Json
$journal = $case.journals | Where-Object operation -eq 'restore'
if ($journal.status -eq 'captured') {
    [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($journal.manifest))
}
```

This display may replace invalid UTF-8 for readability. The base64 value and
its SHA-256 remain the authoritative bytes. An admission record stores its own
fields directly; restore/purge journals also contain a base64 `RecordData`
encoding that must be decoded separately for a complete review.

## Cases that must remain retained

| Uncertainty | Required handling |
| --- | --- |
| An operation directory, manifest, stage or capture exists without a recorded identity | Retain it as unidentified. Do not adopt or delete it merely because its name looks reserved. |
| Admission parent/capture/source identity changed, or rollback destination is occupied | Preserve both objects, journal and any verified snapshot. Do not overwrite the current original path or attach a replacement to the old record. |
| Restore is `publishing`, but the destination is missing or has a different identity | Preserve the destination and all backups. Do not mark the record restored or retire its snapshot using a content match. Publication may have happened before a later rename/replacement. |
| Restore has a valid durable publication/completion receipt but destination was later changed | Preserve the later destination. The receipt proves a historical event; recovery does not restore over that destination. |
| Purge claim/delete state cannot be tied to the approved file identity | Retain the claim, journal and other artifacts. Do not broaden the purge to neighboring or replacement files. |
| Conflicting, malformed or oversized journals | Preserve the entire case and investigate the cause. An inspection hash is not approval to replay an invalid journal. |

## Explicitly hold a reviewed restore case

Version `12.0.103-native` introduced:

```text
utmstack_edr quarantine hold <id> --inspection <saved-json> --reason <text>
```

Perform the capture and maintenance-window steps above first. This command
requires a version-1 inspection for this exact ID, one captured restore journal,
and a nonblank reason of at most 4096 UTF-8 bytes. It rechecks the root,
operation-directory and manifest identities and SHA-256 under the item lock.
A replaced file with identical bytes cannot inherit the approval. It opens
existing protected storage without SQLite, automatic recovery, log creation
or permission changes. The item lock and decision storage are intentional writes.

The durable decision records the exact original bytes and identities, decision
UUID, UTC time, process user UID/SID and reason. The identity names the account
running the command; it does not identify a person operating a shared account.
Preserve its JSON output, diagnostics and exit status with the inspection. A
retry using the same evidence, reason and account returns the existing decision.
A failed command may have retained a preparation or published a fence; preserve
the reported decision ID and inspect again before any further management action.

The decision occupies `.purges/<id>` as a private regular file. Supported
directory-only recovery paths from `12.0.102-native` refuse this entry rather
than treating it as a purge approval or an absent journal. The original restore
journal stays intact. New recovery reports `operator_held`, restore and manual
purge refuse the item, and automatic retention skips a valid hold. A malformed,
replaced or changed decision fails closed as an integrity error. An interrupted
preparation stays inside the original restore operation as an unknown entry;
recovery retains it for investigation. Do not delete or edit either marker.

A hold is an explicit decision to retain evidence. It does not cancel a restore,
prove publication, mark a database record restored/purged, or adopt a destination,
stage or backup. It can preserve malformed original journal bytes without
trusting their claims. Version 103 provides retention only; version 104's reviewed
release is described below. Cancellation and a separately approved fresh restore
remain pending. There is no force-adopt, force-delete or journal-edit command.
Do not remove a hold fence manually. This retention decision does not qualify
Windows hardware power-loss durability. The
[retention-hold audit](../../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-hold/REPORT.md)
records the command's bounded native Windows AMD64 and Linux ARM64 qualification,
earlier-v102 API refusal, process-exit boundaries and retained failures. The
[GCP follow-up](../../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-gcp/REPORT.md)
executes the earlier inspection diagnostic and its CLI main entry point on
Windows Server 2022 AMD64 and Debian 12 ARM64. That earlier inspection run does
not qualify other operating system releases or the later retention command.
Further runtime
campaigns use disposable GCP VMs. Docker is permitted only inside those VMs;
do not start or invoke it on this Mac, including for inventory or cleanup.
Delete the VMs and their task-created supporting resources after testing.

## Reviewed hold release — local v104

Local main at `12.0.104-native` adds an explicit release:

```text
utmstack_edr quarantine release-hold <id> --hold <decision-uuid> --inspection <saved-json> --reason <text>
```

The [release audit](../../../docs/edr/native-engine/audits/2026-11-12-v274-recovery-release/REPORT.md)
records bounded native Windows AMD64/Linux ARM64 and process-exit qualification.
First follow the maintenance-window, backup and inspection steps above. Save a
new inspection that includes both the unchanged original restore
journal and its exact regular-file hold fence. Select the hold UUID explicitly
and supply a new release reason. Releasing this hold authorizes normal recovery
of the original approved operation on a later recovery command or service start.
Review that effect before issuing the command. Release does not itself run
recovery, cancel a restore, approve a fresh operation or prove publication.

The standalone command opens no SQLite database or log and does not harden
storage. Under the item lock it checks the root, original directory/file/hash,
hold UUID, containing directory identity, fence file identity and exact hash.
Changed or malformed bindings are refused. The original restore operation,
destination, snapshots and stage are never moved or edited by release.

Before removing the active fence, the command records an approval with UTC time,
process-account identity, reason, exact hold bytes and all approved identities
in a new private `.reviews/<hold-uuid>/manifest.json`. It flushes that file and
directory, then moves the same fence file, without replacement, from
`.purges/<id>` to `.reviews/<hold-uuid>/hold.json` and syncs both directories.
The fence file identity is preserved. An interrupted preparation without a valid
receipt remains retained; a plausible directory name is never adopted. The
history is bounded to 1024 entries; there is no automatic history cleanup.

Preserve JSON output, diagnostics and exit status. An error can leave an approval
or completed fence move; do not remove markers manually. Retry with the same
saved inspection, hold UUID, reason and account. An exact valid approval can
resume an interrupted move; completed history can return the prior decision
even after normal recovery has retired the original journal. That historical
result applies only to the specified hold and cannot release a later hold for
the same quarantine ID. Changed identities, unknown history children and partial
receipts are retained for investigation. The same fence file present in both
active storage and history is ambiguous and is refused, even when its bytes and
file identity match. A hard link cannot serve as proof of a completed move.

Malformed original journal bytes remain opaque evidence: release allows ordinary
recovery to report their existing integrity failure, not replay them as valid.
Ambiguous publication still needs its original identity proof after release.
Cancellation, archived-stage proof and separately approved fresh restore remain
pending. Six actual process-exit controls and two final forced guest-reset
controls pass
with explicitly synced starting artifacts. Windows directory sync is still the
existing no-op; readiness writes and reset latency also prevent an arbitrary
hardware power-loss claim.

The earlier unflushed raw-source reset remains failed. Read-only diagnosis
confirms that the same raw file ID contains seven NUL bytes while the separate
durable snapshot retains the expected payload. At the historical v104 checkpoint,
no production admission fix had been made. A valid journal or snapshot does not
prove that all raw-source data survived
machine loss; preserve and independently examine those artifacts before further
management. Release only checks its exact journal/fence approval and does not
certify raw-source persistence. The stronger final release fixture precondition
does not close this admission-durability finding.


## Bounded Windows source-flush qualification — November 17

The [source-flush audit](../../../docs/edr/native-engine/audits/2026-11-17-windows-native-source-flush/REPORT.md)
retains a separate successful correction at source `663e4618ce`, package108.
Production admission flushes the pinned source before intent and after capture
protection. Access-denied file flushes alone may use a derived local-volume
barrier after exact volume identity checks. Other errors remain failures; the
original strict source pin and retained-writer refusal stay enforced.

The native Windows source-access matrix passes, followed by two actual guest
resets on eight fresh raw sources whose fixture does not sync input. Source,
journal and directory identity and bytes are verified before recovery. These are
bounded Server2022 AMD64 controls with concurrent volume activity and measured
reset delays. They do not repair the historical v104 artifacts, establish generic
Windows directory sync, certify hardware power loss or approve cancellation,
archive movement, restore publication or a fresh operation. Keep the existing
reviewed hold/release procedure and independently inspect retained evidence.


## Reviewed restore cancellation — candidate source 5cc4ef7227

Cancellation is separate from releasing a hold. Stop the service, retain the case
capture, create the ordinary retention hold, then save a new `quarantine inspect`
result containing both the exact original restore journal and the active regular
fence. Review that saved evidence and the hold decision UUID before running:

```text
utmstack_edr quarantine cancel-restore <id> --hold <decision-uuid> --inspection <saved-held-case-json> --reason <text>
```

This standalone command opens existing private storage without SQLite, logging,
hardening or automatic recovery. It verifies the reviewed root, original
operation directory, original journal identity/hash, fence container, fence
identity/hash and hold UUID under the item lock. The cancellation reason is a
new explicit decision. Save the output and retain the original held-case input.
A lost-return retry requires that same input, hold UUID, reason and actor.
A new inspection after cancellation has an absent active restore journal and
cannot replace the original approval input.

Before moving anything, cancellation writes and flushes a distinct version-1
`restore-cancellation` receipt in `.reviews/<hold-uuid>/manifest.json`. The
receipt pins its own file/directory identities and exact reviewed hold evidence.
It permits the same original operation directory at either `.restores/<id>` or
`.reviews/<hold-uuid>/restore`, never both. The original directory is moved
without replacement, preserving every unknown child. Neither the original
journal nor the unchanged active `.purges/<id>` fence is copied or rewritten.
Source captures, independent snapshots, staged output and any ambiguous original
destination are retained. No restore/purge or historical completion is claimed.

The active version-2 fence plus this distinct cancellation receipt form the new
binding. An older reader refuses the changed original location. Current normal
restore, purge and retention remain fenced. `release-hold` deliberately refuses
cancellation history; it cannot turn cancellation into resumed original recovery.
Separately approved fresh restore and later cancellation-fence release are still
pending and must not be approximated by deleting a journal or fence.

A complete interrupted receipt can be retried after exact identity/byte checks
and another flush. A bare history directory, partial receipt, unknown history
child, copied receipt/fence/original, changed journal, actor or reason is refused
and retained. Receipt bytes are checked again immediately before the directory
move. Unknown children *inside the original operation* are preserved rather than
interpreted or removed. On Windows, the move pins the private directory with
DELETE access and renames that exact handle relative to its pinned destination.

The restore-stage lookup remains an ownership hint. It follows only the exact
quarantine ID, active fence UUID and cancellation archive, corroborating the
original journal and actual stage identity. There is no prefix exemption or
history scan. Startup can rebuild a missing hint from a valid bound original
manifest, including after cancellation; malformed original bytes never establish
stage ownership.

The candidate has native Windows Server2022 AMD64 and Linux AMD64/ARM64 package,
process-interruption and earlier-v104 API evidence. The [frozen audit](../../../docs/edr/native-engine/audits/2026-11-17-restore-cancellation/REPORT.md)
also records two forced guest resets with eight cancellation cases across C:/D:
volumes. Initial raw and unknown fixture bytes were explicitly flushed: these
controls isolate cancellation metadata and do not extend admission claims.
Publication is tracked separately in the Windows handover. Windows directory
sync remains the existing no-op. No arbitrary hardware power-loss, production
signing, wider Windows matrix or full replacement claim follows from this API.

## Read-only fresh-restore review — source 4b0932e2d1

After reviewed cancellation has archived the original operation, keep the
service stopped and preserve the entire case, active hold and cancellation
output. Choose an explicit local destination whose parent already exists and
whose leaf is absent. Then collect an independent observation:

```text
utmstack_edr quarantine inspect-fresh <id> --hold <decision-uuid> --destination <absolute-local-path>
```

Save the exact UTF-8 JSON output, diagnostics and exit status. On Windows, use
native process capture or Command Prompt redirection to preserve stdout bytes;
PowerShell text redirection can change their encoding. Paths and record bytes
use base64 encodings to preserve Unix filenames. Review decoded bytes without
editing the saved JSON.

The command validates the cancellation receipt, same archived original and
unchanged active fence, then independently hashes the retained snapshot and
checks its file identity, size and approved record. It pins the actual new
destination parent identity. It refuses existing files/directories, including
outputs with identical bytes, and destinations inside the retained storage or
original restore stage, even if that stage was renamed. Resolve conflicting
objects independently; do not delete ambiguous evidence to make this check pass.

Inspection neither opens SQLite nor creates locks, logs, stages or destinations.
Its result is an observation, not a lease or publication approval. Ordinary
restore and purge remain held. Fresh execution and later cancellation-fence
release are separate work; do not substitute ordinary `restore`, `release-hold`
or manual journal/fence deletion.

The [frozen inspection audit](../../../docs/edr/native-engine/audits/2026-11-17-fresh-restore-inspection/REPORT.md)
binds all 5516 inputs to this source commit and retains native Windows AMD64 and
Linux AMD64/ARM64 full suites, vet and affected controls. This read-only API
adds no guest-reset, hardware-loss or production-delivery qualification.


## Independently reviewed fresh approval and execution

The following commands have bounded native evidence in the
[execution audit](../../../docs/edr/native-engine/audits/2026-11-17-fresh-restore-execution/REPORT.md),
including Windows AMD64, Linux AMD64/ARM64 and 32 controlled Windows checkpoint
reset cases. This component qualification does not promote a production package.
Preserve the entire case and keep the service stopped while reviewing it.
Separately reviewed original retirement is described below; it grants no deletion permission.

After saving the exact independent `inspect-fresh` JSON, select a normalized new
operation UUID distinct from the original operation and review the snapshot,
record and explicit absent destination. Then run:

```text
utmstack_edr quarantine approve-fresh <id> --operation <new-uuid> --inspection <saved-fresh-json> --reason <text>
utmstack_edr quarantine execute-fresh <id> --approval <saved-approval-json> --reason <text>
```

Save the exact UTF-8 approval stdout before execution. Approval opens no database,
creates no external stage and releases no fence. Execution requires the same
actual actor as that approval. Preserve each reason and input for exact retries;
a complete saved approval is historical evidence, not a lease on current files.
Execution revalidates the original fence/archive, snapshot and destination before
opening the existing database. A copied or stale record, payload/control file
passed as the database, existing destination or changed authority is refused.

Execution creates a new v2 journal, new private stage, immutable ownership record
and same-file owner ticket before copying any valuable output. It publishes only
a newly identified, verified output without replacement. Publication intent alone
cannot prove that a rename happened: ambiguous stage/destination objects remain
retained. A successful retry may use the same new destination identity and proven
stage-source absence; matching bytes alone never authorize adoption.

The result identifies the actual new destination and operation. The cache record
keeps its original-source audit path. Only the new stage/ticket/hint is cleaned;
both original raw and independent snapshot backups, the original cancelled
archive/stage and the unchanged retention fence stay intact. Startup does not
execute this permission, and ordinary restore/purge/release remain blocked even
after fresh success. Preserve saved inputs for a lost-return retry rather than
reapproving a different case.

## Separately reviewed cancellation of an incomplete fresh operation

For a bare/partial preparation, unknown child or ambiguous fresh publication,
collect a separate observation of the exact active new operation:

```text
utmstack_edr quarantine inspect-fresh-cancel <id> --operation <new-operation-uuid>
utmstack_edr quarantine cancel-fresh <id> --operation <new-operation-uuid> --inspection <saved-cancellation-json> --reason <text>
```

Save the inspection before cancellation. Inspection is read-only and opens no
SQLite, log or lock. It captures exact approval/ownership and active operation
identities and bounded raw journal bytes, including a missing or malformed
preparation. A recognizable different operation is refused. Cancellation checks
the saved observation under the item lock and flushes a distinct immutable
`cancellation.json` before moving the same active operation directory into that
approval's `restore` archive. It preserves every unknown child without changing
an external stage, destination, backup, database flag or original retention fence.
It makes no publication or completion claim.

Keep the original inspection, reason and actor for an exact lost-return retry.
A partial cancellation receipt is retained and refused. A later inspection after
the move cannot replace the original saved input. The cancelled approval cannot
resume execution. A subsequent fresh operation needs its own new UUID, independent
inspection and absent destination; retained stages remain protected and cannot
become a new publication parent. Ordinary hold release is not a substitute for
the separately reviewed original retirement described below.

On Windows, new privileged private objects use the already trusted SYSTEM and
Administrators principals. Decision actors remain individual process SIDs; SYSTEM
reading an administrator's stage does not grant it permission to execute that
administrator's approval. The descriptor policy for a nonadministrative service
account is preserved. Existing legacy objects are not silently reowned or given
broader trust. Preserve unreadable objects for independent review.

## Separately reviewed retirement of a cancelled recovery

Keep the service stopped and preserve the complete case. Retirement closes the
original cancellation's active hold while permanently blocking ordinary replay
for that quarantine ID. It does not publish a destination, delete a backup,
change the database disposition, or authorize retention cleanup.

First locate the retained original staging directory by its recorded identity.
An explicit renamed location is accepted only for the same directory. Review
all fresh UUIDs belonging to the case: each must be unused, committed, or
separately cancelled. Active, substituted, or unprovable histories retain the
original hold.

Save the complete inspection before making the new decision:

```text
utmstack_edr quarantine inspect-retirement <id> --hold <original-hold-uuid> --stage <absolute-original-stage>
utmstack_edr quarantine retire-cancelled <id> --inspection <saved-retirement-json> --reason <text>
```

The stage argument names the directory containing the original retained
`snapshot`, rather than the snapshot file. Inspection reads existing storage
without creating a log, database, recovery lock, or directory. It records the
same original hold and cancellation, original journal, independent snapshot,
currently observed raw source, stage identity, and bounded fresh-operation
inventory. Preserve exact JSON bytes when saving output; on Windows Command
Prompt, direct output redirection preserves the program's UTF-8 output.

The second command records a separate immutable decision with the actual actor,
new reason, and exact saved inspection. It completes a ticket in the retained
old stage, publishes a private regular-file guard at `.restores/<id>`, and only
then moves the same original `.purges/<id>` hold into the original cancellation
history as `.reviews/<hold-uuid>/hold.json`. All original and fresh data,
snapshots, outputs, archives, and unknown archive children remain in place.

The regular guard is permanent. Supported older directory-only readers refuse
the occupied restore slot even when their database is stale. Current readers
validate its decision and report `retired` without replaying the old journal.
Both manual and automatic purge remain prohibited. A later fresh operation
cannot reuse this slot; any future deletion requires a separate reviewed
protocol.

Preserve the original saved inspection and exact reason after an interrupted
return. Only the same actor and decision can retry. If that original stage was
renamed after a complete decision, add `--stage <same-renamed-original-stage>`;
this changes only the locator, never the approved identity. An empty or partial
decision is retained and refused. A complete decision may finish only the
missing canonical bytes of its own prepared controls, without truncating or
replacing changed objects. Never remove a partial receipt, guard, or ticket to
force a retry.

Historical ownership is independent of the database: intact stage tickets and
their exact private receipts protect known old and fresh stages even after a
stage rename, a recreated database, or admission through another Store. A
complete old-stage ticket also protects an interrupted prepared transition
before guard publication. Missing or damaged bound controls cause an integrity
refusal. This does not establish ownership after all filesystem locators are
lost or justify adopting arbitrary reserved-looking directories.

Permanent guards consume the shared bounded restore queue. At its 1024-entry
limit, inspection and new preparation refuse without evicting existing guards
or evidence. Do not remove markers to reclaim this quota.

The [retirement audit](../../../docs/edr/native-engine/audits/2026-11-17-cancelled-recovery-retirement/REPORT.md)
records native qualification, source bridges, controlled resets, failures and
limits. Windows directory synchronization remains a no-op; bounded guest reset
results cannot establish arbitrary hardware power-loss durability.
