# Linux permission ownership

Permission modes use an independent process of the running EDR executable. The
service starts `/proc/self/exe __fanotify-broker` before activating its group.
The internal entry point runs before service, logger, engine and database startup.
It never executes helpers or scans content. Notification mode keeps the local
monitor. The local monitor still refuses executable/helper mounts in permission
mode; only the independently launched owner can mark those mounts.

## Authority and communication

The parent creates a close-on-exec Unix `SOCK_SEQPACKET` socket pair. Only the
owner endpoint is inherited as descriptor 3. There is no listener or public socket
path. The child verifies that the peer is its same-user parent before activation.
It exclusively owns the fanotify group, original event descriptors, monotonic
deadlines, admission limits, identity rechecks and permission responses. The group
descriptor is never transferred. The scanner receives only a duplicate of the
pinned input through `SCM_RIGHTS`; it cannot substitute a pathname or respond to
the kernel directly. Received descriptors are close-on-exec.

Private version-1 JSON packets are limited to 16 KiB and reject unknown fields,
trailing data, incompatible versions, unexpected descriptors and truncated
ancillary data. All received rights are closed on refusal. Event paths use a byte
encoding, preserving non-UTF-8 Unix names. Configured scope paths must be UTF-8,
and their encoded configuration must fit the packet bound; larger configurations
are refused before starting an owner. Scan replies carry a request ID and the
engine generation. An expired or unknown request cannot authorize a new event.

All socket I/O is nonblocking. Waiting for a writable socket or another writer is
subject to the caller's deadline. Fixed owner and scanner worker pools, bounded
queues and single-use objects prevent a stalled callback from creating unlimited
replacement workers. Observers are separate from permission resolution. Late
detections are reported without a second kernel response.

## Directory scope and ambiguous names

Linux descriptor paths append ` (deleted)` for an unlinked name. A literal
filename with that suffix has the same spelling. The backend retains a bounded
absolute spelling while returning a specific metadata error. Because the marker
affects only the final component, an otherwise canonical path outside every
configured directory root can still be excluded. Such excluded requests are
allowed without running a content scanner or degrading protected coverage.
Sibling prefixes are not children of a configured root.

Inside a protected root, the uncertain name remains incomplete: the configured
failure policy applies and health degrades. It is never supplied as trusted
filename metadata to the engine. A `/` directory root excludes none of these
absolute paths. Other inspection errors, relative/noncanonical/oversized paths
and missing names retain the existing failure behavior. Scope validation at
startup does not qualify hostile directory/mount changes afterward.

## Failure and shutdown

The owner applies configured fail-open/fail-closed behavior to incomplete scans,
deadlines and admission failure while its group remains available. A confirmed
detection is denied. An unchanged complete clean decision is allowed. The scanner
process can stall, enter GC, stop or exit without owning the permission deadline.
The owner never reads exec-locked process image or command-line metadata.

The parent allows five seconds for initial owner readiness, then requires
periodic status within one second. It treats missing status, child exit,
protocol failure and observation overflow as degraded protection. It closes the
channel on cancellation, waits for the child, and forcibly terminates and reaps a
child that fails to exit within one second. Owner status is sticky for observed
scan/coverage errors. Kernel activation errors are returned across the channel.
An individual delayed heartbeat is dropped without closing the group. A lost
outcome increments the observation-loss counter and degrades health while the
independent loop continues to resolve permissions.

Channel EOF tracks parent-process exit or exec. Do not replace it with
`Pdeathsig`: Go documents that this signal tracks the creating OS thread, whose
exit can precede process exit. Parent loss cancels the owner's loop; it resolves
already-read requests according to shutdown policy and closes the group. Closing
or killing an owner releases unread kernel requests as allowed. **FailClosed is
not persistent enforcement after owner/group loss.** A stopped owner is detected
through missing status and eventually killed; its scheduling is not covered by
the per-scan deadline.

## Qualification boundaries

Tests require an explicitly opted-in isolated guest and owned inert fixtures.
Executable/root-mount tests require an additional opt-in. Test scanner callbacks
are separate from the actual foreground-service fixture, which uses the real
engine and event spool. Preserve those distinctions when reporting results.

A mount mark does not cover nested mounts automatically. Root scopes must reside
on declared mount identities. This architecture does not establish every kernel,
filesystem, namespace or security-module profile, arbitrary engine/update overlap,
sustained overload, an installed-service support matrix, power-loss durability, or fleet rollout.
The service's signed release path must carry an advanced distribution version
before production delivery; development acceptance does not publish a release.

References: [fanotify semantics](https://man7.org/linux/man-pages/man7/fanotify.7.html)
and [Go process attributes](https://pkg.go.dev/syscall#SysProcAttr).
