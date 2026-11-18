# Floating-point TIFF controls

Generated for native-go-v251 by the retained `generate.cpp`, `oracle.py`,
`mutate.py` and `generate_float_oracle.py` in the v251 audit directory.

The independent installed libtiff 4.7.2 encoder writes 264 RGB/RGBA32F TIFFs
with both byte orders, strips/tiles, raw/LZW/Deflate/PackBits/ZSTD, applicable
predictors 1/2/3 and three dimensions. Pre-encoding `.samples` files retain
little-endian IEEE bit patterns including signed zeros, subnormals, NaNs,
infinities, negatives, values above one and ordinary fractional colors.
Independent NumPy f32 multiplication/clamping plus rounding produces RGB8
expected bytes; the Go implementation generates none of those expectations.
All 72 ZSTD images are rejected by the pinned native dependency and do not
count as accepted RGB/detection evidence.

Independent Python metadata mutations and raw predictor writers exercise
sample metadata, advisory encoded lengths, physical EOF/padding and zlib
trailer distinctions. Four actual gray-alpha 8/16-bit, white/black TIFFs are
rejected by the locked decoder despite the outer adapter listing GrayA.
The final 307-case manifest pins 210 accepted exact native hashes and 97
native rejections. Every accepted case also has independent RGB8 bytes;
expected RGB files for rejected cases are not acceptance evidence.

Separate binary tables retain 34,057 independent float32 conversions around
every RGB8 rounding boundary and deterministic random bit patterns. These
are finite controls, not exhaustive testing of all 2^32 float patterns.
Builtin-only native and Go engine checks make matching fuzzy rules reachable
without a custom type overlay. These inert samples do not measure malware
recall, fleet false-positive rates or universal TIFF/ClamAV equivalence.
