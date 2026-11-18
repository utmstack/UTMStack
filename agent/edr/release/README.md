# Signed native EDR release archives

`releasepack` creates a ZIP containing only `release.json`, the native Go
service executable and, for Windows, `utmstack_amsi.dll`. The manifest is
canonical JSON signed with Ed25519 and binds exact version, platform, names,
sizes and SHA-256 hashes. `release.Extract` checks the signed manifest and
every extracted byte before returning a private staging directory.

After Windows Authenticode signing, a publisher can run, from `agent/`:

```sh
go run ./edr/cmd/releasepack \
  -os windows -arch amd64 \
  -service utmstack_edr_windows_amd64.exe \
  -provider utmstack_amsi.dll \
  -source /secure/release-input \
  -out /secure/release-output/utmstack_edr_windows_amd64.release.zip \
  -key /secure/secret/edr-ed25519-private.b64
```

The key file contains the base64 encoding of the 64-byte Ed25519 private
key. Keep it in the publisher's secret store, never in the repository,
artifact or endpoint. The tool fails if the output archive already exists.
The public key printed by the tool must be pinned into the agent release by
a separate authenticated build step. `releasepack` does not perform or verify
Authenticode signing; production Windows release CI must validate both PE
signatures before packing.

The dependency updater now consumes this signed ZIP and transactionally
promotes the service, provider and manifest under a build-time pinned public
key. Production release CI has **not yet** provisioned the matching private
key, signed the Windows provider, published these archives or proved native
service/AMSI readiness and crash recovery. Until those gates pass, this
format is an implementation path rather than an approved production release.

On Linux, release readiness also checks locally enabled fanotify protection.
The fresh exact-executable status must report the access monitor active, and
permission modes must additionally report active file prevention. Unsupported,
degraded, missing or disabled requested protection prevents commit and invokes
the existing rollback protocol. Intentionally disabled/default-off fanotify
retains its existing behavior. Malformed, oversized or nonregular local
configuration fails readiness rather than blocking the updater on a FIFO.

Explicitly enabled Linux network enforcement also requires a supported,
requested active `network_enforcement` state without an error. A cold cache
reconciliation or unresolved management safety policy cannot be committed as a
healthy release. Disabled/detect-only networking stays optional. The readiness
projection honors the configuration loader's default-on blocklist values when
those fields are absent.

A positively observed systemd `LoadState=loaded`, `ActiveState=failed` unit
remains an installed stopped service for recovery. Other service errors remain
errors; this does not unmask or adopt an absent/broken unit. The Linux native
acceptance entrypoints in `dependency/edr_installed_linux_test.go` require an
explicitly owned isolated root guest, an exact install directory and read-only
publisher inputs. Ordinary package tests cannot install or stop a host service.

November 11/12 v274 laboratory acceptance now proves actual Windows AMD64 SCM,
registered owned/system COM/AMSI, ordinary-user published access, two distinct
signed pairs, rejection before stop, readiness rollback, abrupt child exit with
repeated journal recovery, and offline restart. Windows staging uses protected
native DACLs; publication refuses retained writers/deleters, and replacement
waits for the pinned service process to exit. See the
[v274 audit](../../../docs/edr/native-engine/audits/2026-11-11-v274-windows-native/REPORT.md)
for exact scope and retained failures. These disposable keys/certificates do not
qualify production custody, real machine power loss or fleet delivery.
