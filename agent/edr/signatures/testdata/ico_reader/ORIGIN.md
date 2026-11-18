# Extended ICO reader controls

The 86 inert files were independently encoded with Python `struct`,
`binascii`, and `zlib` using ICO, PNG, and DIB field layouts. They contain
no executable code. The generator is retained at
`docs/edr/native-engine/audits/2026-11-07-v247-ico-reader/generate_oracle.py`.

The SHA-pinned ClamAV 1.5.4 `sigtool --fuzzy-img` supplies the expected
hash or decoder rejection for each file. `manifest.json` records all input
sizes and SHA-256 values. The tool's SHA-256 is
`cac0f0dcca225dd058246df6f142de5860627e0f6eeee1969d0dcfbd4cb7e2df`.

There are 67 accepted images and 19 rejected images. Controls include
advisory directory lengths; Adam7, RGB/palette transparency, and unsupported
PNG output color types; CORE and INFO bitmaps; short palettes with unused
indices; top-down and odd heights; RLE4/RLE8 commands, cursor/AND-mask
boundaries, overruns, no-op deltas and missing terminators; and image sizes
larger than the directory's 256-pixel fields.

These generated controls establish selected decoder/transform behavior.
They are not a real-world malware or broad benign population, or an
exhaustive image-format proof.
