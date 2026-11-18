# Fixed image-hash cosine transform

The Go 32-point f32 DCT-II is adapted from RustDCT 0.7.1's type-II
butterflies (2, 4, 8, 16), the 32-point split-radix plan, and twiddle
construction. Copyright (c) 2015 The RustFFT Developers.

The upstream crate is dual MIT/Apache-2.0 licensed. This adaptation uses
its MIT option; the full notice and permission are in `LICENSE.MIT`.
The pinned archive has SHA-256
`8b61555105d6a9bf98797c063c362a1d24ed8ab0431655e38f1cf51e52089551`.

Source: https://crates.io/crates/rustdct/0.7.1

Only the fixed transform used for image hashing is included. The Go
implementation has fixed stack buffers, bounded slice accesses and explicit
f32 product rounding. It does not invoke Rust, ClamAV, an FFT library or a
native executable at runtime.
