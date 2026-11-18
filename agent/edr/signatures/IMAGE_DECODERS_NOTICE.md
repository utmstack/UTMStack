# HDR, DDS and TIFF reader references

The Go HDR and DDS/DXT readers use image-rs 0.25.9 as their behavioral and
algorithmic reference, including HDR scanline/RGBE conversion, DDS header
admission and DXT color interpolation. The TIFF profile follows its CMYK8/16
f32 conversion and supported color/sample admission. The private IFD view,
physical guards and integration with the existing Go TIFF codec are owned Go
code. The project source is dual
MIT/Apache-2.0 licensed. This adaptation uses the MIT option; its complete
permission notice is retained in `IMAGE_DECODERS_LICENSE.MIT`.

Source: https://github.com/image-rs/image/tree/v0.25.9/src/codecs

The owned floating-point TIFF chunk reader follows image-tiff 0.10.3's
horizontal and byte-plane prediction, chunk/padding and streaming semantics.
That reference uses the MIT license, copyright 2018 PistonDevelopers; its
complete notice is retained in `TIFF_SAMPLES_LICENSE.MIT`.
Source: https://github.com/image-rs/image-tiff/tree/v0.10.3/src/decoder
Raw samples are decoded in Go; LZW uses the existing Go TIFF LZW package,
and Deflate uses Go's standard zlib package.

The owned JPEG-compressed TIFF reader follows the same TIFF reference for
table reconstruction, physical chunk bounds and sample/predictor meaning.
It uses the Go `jpeg154` raw-component profile; see that package's MIT/BSD
notices for its algorithm sources. No laboratory SDK is a runtime dependency.

The Go implementation decodes bounded physical views, checks scan work
and cancellation, and computes RGB directly because the fuzzy image hash
discards alpha. It invokes no image-rs, Rust, ClamAV or native executable
at runtime.

The Lanczos3 resizer uses an owned Go port of the bounded scalar Arm sine
polynomial, from optimized-routines commit
20859d3fd9e797dde592c489c8ef7d383f778620, copyright 2018–2026 Arm Limited.
This adaptation uses the MIT option; the upstream license is retained in
`LICENSE.resize-math`. Source:
https://github.com/ARM-software/optimized-routines/tree/20859d3fd9e797dde592c489c8ef7d383f778620/math

`resize_math_macos.bin` contains independently measured numerical corrections
for the Mac reference. The v263 image-resize audit retains all generation
inputs, the complete finite geometry sweep, and independent native and Go
validation. The table is independent of image content and contains no malware
signatures. Mac and Linux hash disagreements remain uncertain. These two
qualified math profiles do not establish parity with every other platform's
math library or expand supported decoder formats.
