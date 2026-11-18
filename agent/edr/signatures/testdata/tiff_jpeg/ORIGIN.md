# JPEG-compressed TIFF oracle

Generated for native-go-v252 by the retained C++ encoders and Python writers
in the v252 audit directory. Installed libtiff 4.7.2/libjpeg produce 256
JPEG-compressed grayscale/RGB/CMYK/YCbCr TIFFs across byte order, strips/tiles,
shared table modes, quality and dimensions. Pillow supplies baseline and
progressive JPEGs. A separate libjpeg encoder supplies 4:1:1 and 4:1:0 with
baseline/progressive and restart variations. Independent metadata writers
exercise table boundaries, EOF, predictors, sample reinterpretation and
photometric handling.

The exact image-tiff JPEG dependency, zune-jpeg 0.4.21 with zune-core 0.4.12,
is compiled into a task-owned Rust laboratory oracle from checksum-verified
crates. It emits raw interleaved component bytes using the same colorspace
option as the locked TIFF adapter. Independent Python/NumPy code assembles
those buffers into expected RGB8 TIFF pixels. No Go code creates expected
pixels. Raw streams and their outputs are retained in jpeg154/testdata/components.

The final 323-file manifest pins 240 accepted native ClamAV 1.5.4 fuzzy
hashes and 83 rejections. All 240 accepted cases have independent RGB bytes.
273 distinct embedded JPEG streams have 262 accepted exact raw-buffer
oracles and eleven refusals. Rejected inputs with available expected pixels
do not count as accepted reader evidence. Generic progressive sampling with
restart remains an explicitly unqualified/refused profile. These inert
controls do not establish malware efficacy or universal TIFF/JPEG equivalence.
