# Locked TIFF JPEG component oracles

Inputs are independently encoded libtiff/libjpeg/Pillow JPEG streams,
including table reconstruction and metadata mutations assembled by the v252
Python laboratory tooling. Output `.raw` bytes come only from the exact
checksum-verified zune-jpeg 0.4.21 / zune-core 0.4.12 code compiled with the
isolated official Rust 1.90.0 SDK. SDK binaries and Rust code are laboratory
tools only; production decoding runs in Go.

The manifest pins 273 streams, 262 accepted raw buffers and eleven native
refusals. JPEG labels describe the older decoder's inferred input space;
TIFF interpretation comes from its own metadata. No color conversion or
Adobe CMYK inversion is applied to raw channels. Standalone-manifest controls
also check missing EOI and required quantization tables against the pinned
ClamAV standalone decoder. See the v252 report for limitations and identity
receipts. Golden bytes never come from the Go implementation under test.
