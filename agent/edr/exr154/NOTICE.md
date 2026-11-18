# OpenEXR image decoder references

This owned Go implementation follows the ClamAV 1.5.4 image profile and
adapts the decoding algorithms of exrs 1.74.0, including RLE/ZIP prediction,
PXR24, B44 and PIZ. The archive is verified against ClamAV's Cargo.lock SHA-256:
`4300e043a56aa2cb633c01af81ca8f699a321879a7854d3896a0ba89056363be`.

The exrs source credits the OpenEXR project for the underlying codecs.
OpenEXR credits Christian Rouet for its PIZ Huffman routines and Loren
Carpenter/Pixar for PXR24. Both projects' copyright, BSD-3-Clause conditions and disclaimers are
retained verbatim in `LICENSE.BSD.md`. The deterministic B44 log table is
retained as little-endian 16-bit entries in `b44_log.bin` from exrs's
`src/compression/b44/table.rs`, itself generated from OpenEXR's table.

The display/data-window and RGB float conversion profile follows image-rs
0.25.9. Its MIT option and full permission notice are retained in
`LICENSE.image-MIT`.

Public references:
- https://crates.io/crates/exr/1.74.0
- https://github.com/AcademySoftwareFoundation/openexr
- https://github.com/image-rs/image/blob/v0.25.9/src/codecs/openexr.rs

The runtime reads bounded in-memory bytes and uses Go code and embedded
data only. Native OpenEXR and ClamAV executables are independent laboratory
encoders/oracles and are not runtime dependencies.
