# HDR and DDS native reference controls

These 178 inert images and malformed controls are independently written by
`docs/edr/native-engine/audits/2026-11-07-v248-hdr-dds/generate_oracle.py`.
`manifest.json` pins their bytes, SHA-256 and ClamAV 1.5.4 sigtool hash, or
null when the reference refuses to hash the input. The reference executable
and libraries are pinned in that audit's `reference-inputs.json`.

There are 146 accepted and 32 rejected controls. They cover raw RGBE, both
HDR RLE families, metadata float grammar, scanline bounds and orientation;
DDS DXT1/3/5, BC1/2/3 DX10 formats, header flags, dimensions, resource/alpha
metadata and physical truncation. Shapes include 4x4, 512x36 and 4x512.
HDR exposure/color/aspect metadata is parsed but does not alter hash pixels,
as in the pinned reader. DDS alpha is discarded after validating encoded
block bounds; only the first image/mip contributes.

Twelve accepted controls retain resize uncertainty: five tall HDRs, three
tall DXT1s and two each of the tall DXT3/DXT5 variants. The tests require
all certain bits to match and check an incomplete matching rule plus a
definite different-hash exclusion. The other 134 have exact hash/rule
results. Rejection is not overall scanner clean/incomplete equivalence.

`resize-independent.ppm` is an additional independent RGB8 conversion of
`hdr-4x512-raw.hdr`, outside the 178-row manifest. SHA-256:
`7693216effdf3605a51f5ab77c52168906386e5f0b89be2005ef3e13290e9849`.
It reproduces a native sinf-dependent half-integer resize disagreement.
Both the reference hash and the earlier Go hash must remain incomplete,
preventing confident false decisions. The audit retains the conversion,
local C comparison and exact native/Go resized PGM pixels.

Native positive rules use explicit harmless local `.ftm` image typing.
Without that overlay, both tested default classifiers keep the HDR/DDS
controls as binary data and do not run fuzzy image rules. These controls
establish reader paths, not official-feed malware recall or complete
image decoder/resize parity. They execute no sample payloads.
