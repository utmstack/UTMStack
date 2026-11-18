# ICO bitmap decoder controls

These 118 inert ICOs were independently encoded with Python's `struct`
module using ICO/DIB field layouts. They contain binary channel extrema or
varied intermediate color values. They do not contain executable code.

The generator is retained at
`docs/edr/native-engine/audits/2026-11-07-v246-ico-directory/generate_bitfields.py`.
It invokes the pinned ClamAV 1.5.4 `sigtool --fuzzy-img` executable
(SHA-256 `cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df`)
to obtain each expected image hash. `manifest.json` records each input's
size, SHA-256, encoding parameters and native hash. The scanner additionally
checks both matching and different fuzzy rules in the optional native tests.

Coverage includes extended BI_RGB headers at 16/24/32 bits; ignored extension
fields; and BI_BITFIELDS RGB555, RGB565, ARGB4444, RGB1010102, RGB777, RGB123
and reversed BGR888 layouts with 40/52/56/108/124-byte headers. V4/V5 bitfield
fixtures include the extra 12 bytes skipped by the pinned image-rs reader.
All fixtures are 32 by 32 pixels. This is decoder acceptance evidence, not
a representative image population or an exhaustive bit-pattern proof.
