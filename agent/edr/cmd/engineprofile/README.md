# Offline engine memory diagnostic

This development command starts no service or profile listener and never
executes, quarantines or modifies a scanned input. It writes only its new private
output directory. Use pinned local signature files and safe fixtures.

```sh
CGO_ENABLED=0 go build -o /absolute/lab/engineprofile ./edr/cmd/engineprofile
/absolute/lab/engineprofile -db /absolute/lab/pinned-db \
  -out /absolute/lab/new-profile-directory -reload
go tool pprof -top /absolute/lab/engineprofile \
  /absolute/lab/new-profile-directory/loaded.heap.pprof
go tool pprof -top -alloc_space /absolute/lab/engineprofile \
  /absolute/lab/new-profile-directory/allocations.pprof
```

Optional positional file paths are bounded by the selected engine file limit.
`-memory-limit-mib 2048` changes Go's process-global **soft** memory limit only
inside this diagnostic process. It can increase GC CPU/time and is not an RSS
cap. It is not a production EDR configuration field.

`phases.jsonl` records exact runtime counters at phase boundaries plus observed
heap peaks from a nominal 5 ms timer. It deliberately forces collection for
retained profiles and explicitly releases memory after stop. Sampling can miss
brief peaks, and these operations change GC timing: do not substitute this
instrumented run for an uninstrumented process-RSS or service benchmark. Compare
identical database bytes/build settings and retain the executable hash.

The reload experiment retains the current generation during compilation but
does not hold long-running scans on it. Signed activation/local-overlay paths
can compile additional databases; measure those separately.
