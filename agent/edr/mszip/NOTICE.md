# MSZIP decoder reference and license

The Go MSZIP implementation is adapted from libmspack's inflater and repair
policy by Stuart Caie. Copyright (C) 2003–2023 Stuart Caie. It is distributed
under the GNU Lesser General Public License, version 2.1; see
[LICENSE.LGPL-2.1](LICENSE.LGPL-2.1). Canonical Huffman table construction in
the reference credits David Tritscher; this adaptation uses canonical counts
and symbol lists rather than the reference's table allocation machinery.

The reference is ClamAV 1.5.4's bundled
[`mszipd.c`](https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/libclammspack/mspack/mszipd.c),
SHA-256 `04ba76766f723552df4dc35b455ccdf6f6a795d13a549f68a314a2bae1d277ba`.
Its `readbits.h` and `readhuff.h` supply lookahead and complete-tree semantics.
Changes replace C pointers, callbacks and decode tables with bounded Go
storage, cancellation, resource-limit aborts and an `io.Reader` input.
The CAB adapter controls record validation and aggregate input limits.
No native library or external process participates in production decoding.
