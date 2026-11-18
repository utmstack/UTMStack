# JPEG decoder source

`reader.go`, `huffman.go`, and `scan.go` are adapted from Go 1.25.7's
`image/jpeg` decoder (The Go Authors). Their BSD license is
[`LICENSE.GO-BSD`](LICENSE.GO-BSD). This package omits the encoder and global
image-format registration. The YCbCr-to-RGBA drawing call uses `image/draw`
in place of Go's internal helper.

`idct_zune.go` is a Go port of the scalar inverse DCT in `zune-jpeg` 0.5.5
(zune-image developers), selected under the MIT license in
[`LICENSE.ZUNE-MIT`](LICENSE.ZUNE-MIT). The fetched crate archive had SHA-256
`dc6fb7703e32e9a07fb3f757360338b3a567a5054f21b5f52a666752e333d58e`.
The 3:1 subsampled CMYK interpolation and 8-bit color multiplication also
follow that crate's MIT-licensed algorithms.

This decoder is part of the Go engine. It does not link to or invoke ClamAV or
Rust at runtime. The copied Go files retain their original BSD notices.

`components.go` adds the TIFF raw-channel profile from the separately locked
`zune-jpeg` 0.4.21 dependency, including its horizontal edge and generic
sampling behavior. Its archive SHA-256 is
`29ce2c8a9384ad323cf564b67da86e21d3cfdff87908bc1223ed5c99bc792713`.
That reference is selected under the same MIT license. Pinned Rust code is
used only by the independent laboratory pixel oracle, not this Go package.
Scan/reconstruction/output work callbacks and pre-allocation dimension caps
bound the new component path. Required quantization-table checks and the
qualified missing-EOI behavior also apply to standalone JPEG decoding.

The TIFF profile also retains independent per-component sampling and padded
strides, including the older dependency's incremental stride setup and
progressive non-luma scan traversal. Native-panic sampling profiles are
returned as unsupported inputs rather than reproduced as panics in Go.
These changes use the same pinned 0.4.21 MIT-licensed reference; standalone
JPEG output keeps its separately qualified sampling model.
