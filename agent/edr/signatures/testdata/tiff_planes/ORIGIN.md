# Independent TIFF component-plane controls

The 200 inert TIFFs wrap the independently libjpeg-encoded JPEG streams in
`agent/edr/jpeg154/testdata/planes`. Three-channel streams use RGB TIFF
metadata; four-channel streams are interpreted as CMYK or RGBA. The TIFF
adapter consumes the raw component byte stream rather than applying JPEG's
standalone display color conversion.

`docs/edr/native-engine/audits/2026-11-07-v253-jpeg-planes/oracle.py` writes
the metadata using the prior independent TIFF writer, invokes the pinned
ClamAV 1.5.4 `sigtool --fuzzy-img`, and independently converts Rust-oracle
samples to expected RGB with NumPy float32 CMYK arithmetic where needed.
Every one of the 114 native-positive inputs has independent `.rgb` bytes;
86 have verified native reader refusals. Identities and reference errors
are in `manifest.json`. No Go output produces expected hashes or pixels.

The v253 audit records source/compiler/encoder identities, red controls,
positive and different-hash native rules, actual Go engine pipeline checks,
budget/cancellation checks and bounded fuzz runs. Test scope is finite and
must not be reported as full ClamAV replacement or malware efficacy.
