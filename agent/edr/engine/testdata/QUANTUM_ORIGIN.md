# Quantum CAB regression archive

`libmspack_mszip_lzx_qtm.cab` is the 379-byte mixed-compression regression
archive from
[`libmspack/test/test_files/cabd/mszip_lzx_qtm.cab`](https://github.com/kyz/libmspack/blob/55d501976171397ccd5d5a7a1ca7da065b1d9a06/libmspack/test/test_files/cabd/mszip_lzx_qtm.cab).
It contains three short harmless text files compressed with MSZIP, LZX and
Quantum. SHA-256:
`0ce0b55fe705b744d41bb361170c0467db30da0c7f9bdd386d5dade71a78e171`.

The source repository is libmspack by Stuart Caie, distributed under LGPL
2.1. The decoder's [notice](../../quantum/NOTICE.md) and
[license](../../quantum/LICENSE.LGPL-2.1) accompany this regression input.
ClamAV 1.5.4 independently extracts the 59-byte Quantum member with SHA-256
`bdcfdaf09e54d61f950b165b201d4ad5f5acfdecff1fc5641e382aa382c74b45`.
The cabinet and its contents are scanned as data and never executed.

The `quantum/` directory also contains three historical libmspack regression
inputs at the same pinned commit:

| File | Source path | SHA-256 |
| --- | --- | --- |
| `cve-2010-2801-qtm-flush.cab` | `cabextract/test/bugs/cve-2010-2801-qtm-flush.cab` | `78bb4c9540e36d3d4228e22cdff0a5677f9c2dddc4239c322277d0b9125d8364` |
| `cve-2014-9556-qtm-infinite-loop.cab` | `libmspack/test/test_files/cabd/cve-2014-9556-qtm-infinite-loop.cab` | `2e6f57d26e12f3f1c0f1f968d7ebb92e77ea7720ea6f9e18f11eb00ef0e845ad` |
| `cve-2018-18584-qtm-max-size-block.cab` | `libmspack/test/test_files/cabd/cve-2018-18584-qtm-max-size-block.cab` | `3ba0c719d8fb12dc531e2a7790c2d15c12d8840de8e4ddc60572c4145181fbf4` |

The 2010 flush control is valid: both Go and ClamAV decode its 16 frames to
524,159 zero bytes. The other two remain malformed/incomplete in Go.
The 18 additional `.cab` files and `manifest.json` are generated harmless
controls from the [retained encoder](../../../../docs/edr/native-engine/audits/2026-11-07-v241-quantum/encode_quantum.py).
Their exact decoded hashes are independently checked against ClamAV 1.5.4.
