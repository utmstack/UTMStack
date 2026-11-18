# TIFF CMYK and reference-admission controls

Generated for native-go-v250 by the retained `generate.cpp` and `oracle.py`
in `docs/edr/native-engine/audits/2026-11-07-v250-tiff-cmyk`.

The independent libtiff 4.7.2 encoder writes 144 CMYK8/16 TIFFs spanning
little/big byte order, strips and padded tiles, raw/LZW/Deflate/PackBits,
horizontal predictors for LZW/Deflate, and three image dimensions. Original
little-endian `.samples` files precede SDK compression. Independent NumPy
f32 arithmetic produces RGB8 goldens, truncating at the original depth
before RGB16-to-RGB8 rounding. The Go implementation does not generate
these expected samples or RGB bytes. The two RGB controls use Pillow.

The 159-case manifest pins TIFF bytes and native ClamAV 1.5.4 hashes or
rejection. All 146 accepted controls have pinned independent RGB bytes.
The three palette names denote 2/16/256-color quantization; Pillow writes
8-bit palette storage. Thirteen rejected controls cover palette, planar/sample/color/depth and
an EOF pixel pointer into a potential virtual IFD. The last control caught
and fixed a prototype false hash; physical source bounds must be enforced
before synthesizing private metadata.

Native rule comparisons use explicit harmless TIFF typing; separate native
builtin-only controls and the engine pipeline also detect CMYK without that
overlay. These are inert rule and pixel controls, not labelled malware or a
claim of universal TIFF/ClamAV equivalence. Float RGB, gray-alpha, JPEG and
other reference-only TIFF profiles remain follow-up work.
