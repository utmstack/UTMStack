# Historical LH4–LH7 regression archives

These six inert `.lzh` archives were copied from the
[Lhasa test collection](https://github.com/fragglet/lhasa/tree/master/test/archives),
distributed under its [ISC license](https://github.com/fragglet/lhasa/blob/master/COPYING.md).
They contain the GPL-2 license text or a benign long test text.

| Local file | Source path | SHA-256 |
| --- | --- | --- |
| `lh5-lha213.lzh` | `test/archives/lha213/lh5.lzh` | `7f10d0be69536733217e984e4ff81d8980f2df06822581d7c006257bfad34851` |
| `lh5-lha213-long.lzh` | `test/archives/lha213/lh5_long.lzh` | `8bc9c2c20bf12d0fa50a8ffbfae5e2fc7af88f63fc3be894fa2aff633db37877` |
| `lh6-amiga.lzh` | `test/archives/lha_amiga_212/lh6.lzh` | `1910a6d4dfe5db3b1ceaf0fb6f2ec5f245d084cc5262c04777c1889e6791afb0` |
| `lh4-amiga-long.lzh` | `test/archives/lha_amiga_122/lh4_long.lzh` | `c48af76b5b2e9498fbf5cf26b8c3013637ca7bd10dea9a39df1a2be52714eef5` |
| `lh6-unix-long.lzh` | `test/archives/lha_unix114i/lh6_long.lzh` | `8f00d1a3197a193ed014054f0b2756930f60163c88b6f5cae8dbc3ecb26e77ae` |
| `lh7-unix-long.lzh` | `test/archives/lha_unix114i/lh7_long.lzh` | `37ffaebe08c9d8dcf8e9add232187bcdaa1313b13f7c22efe810b8bf883eaae5` |

The [v141 audit](../../../../docs/edr/native-engine/audits/2026-10-28-v141-lh1/REPORT.md)
contains Go and pinned ClamAV child-only rule results. Two other Lhasa
archives from LHARK-A v0.4d advertise `-lh7-` but use a distinct LK7
bitstream; they are intentionally excluded from these six standard
method fixtures. The pinned ClamAV scanner did not detect the same child
rule in those two archives, and Go reports incomplete on them.
