# Quantum decoder reference and license

The Go arithmetic-model decoder in this package is adapted from the Quantum
algorithm implemented in libmspack by Stuart Caie, based on Matthew Russotto's
research and implementation (used with permission by libmspack).

Copyright (C) 2003–2023 Stuart Caie. The adaptation is distributed under the
GNU Lesser General Public License, version 2.1; see
[LICENSE.LGPL-2.1](LICENSE.LGPL-2.1). Changes replace libmspack's C I/O and
window machinery with bounded Go slice decoding, context cancellation,
virtual CAB block trailers and the ClamAV adapter's zeroed initial history. No native
library or external process is used by the production decoder.

The reference is
[`qtmd.c`](https://github.com/kyz/libmspack/blob/55d501976171397ccd5d5a7a1ca7da065b1d9a06/libmspack/mspack/qtmd.c)
at commit `55d501976171397ccd5d5a7a1ca7da065b1d9a06`. Source SHA-256:
`cf73514083864ab464fec0ec0600ddce3f7636364531a93e1e66c4eee76e1c41`.
The model tables are generated from their slot formulas rather than copied
as static arrays. The same repository's CAB regression file
`libmspack/test/test_files/cabd/mszip_lzx_qtm.cab` supplies an independent
literal/match control; its attribution is in the engine testdata directory.
