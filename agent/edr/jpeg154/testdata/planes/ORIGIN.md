# Independent TIFF JPEG sampling controls

The 144 JPEGs are inert pattern images encoded by installed libjpeg through
`docs/edr/native-engine/audits/2026-11-07-v253-jpeg-planes/generate.cpp`.
Eighteen three/four-component sampling layouts each cover 37x43 and 32x32
pixels, baseline/progressive scans and restart intervals zero/two. Channels
have unequal sampling, non-first or split maxima, or vertical factor four.

The separate raw-pixel oracle compiles checksum-verified zune-jpeg 0.4.21 /
zune-core 0.4.12, matching ClamAV 1.5.4's locked TIFF dependency, with the
isolated official Rust 1.90.0 SDK. It selects the inferred input colorspace
as its output. It is not the newer standalone JPEG dependency. All 82
accepted buffers have independent `.raw` bytes and identities in
`manifest.json`; 62 failures retain the oracle exit and error. Panics from
this reference are recorded as refusals and independently checked through
ClamAV's TIFF FFI adapter, never reproduced as Go panics.

Expected bytes are produced solely by the independent Rust oracle. The Go
implementation is never used to encode fixtures or establish golden pixels.
SDK/download/extracted-source/encoder identities, generation, red controls
and native hash/rule checks are retained in the v253 audit. These finite
controls establish their own results, not universal sampling or malformed
JPEG compatibility. The production decoder uses Go and does not load these
laboratory tools.
