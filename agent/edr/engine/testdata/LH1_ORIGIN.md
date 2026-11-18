# LH1 regression archives

The eight `.lzh` fixtures here were copied from the
[Lhasa test archive collection](https://github.com/fragglet/lhasa/tree/master/test/archives),
whose repository is distributed under the [ISC license](https://github.com/fragglet/lhasa/blob/master/COPYING.md).
They are historical output of LHArc, MorphOS LHA, Amiga LHA, OS-9 LHA and
OSK LHA;
the compressed members contain the redistributable GPL-2 license text
or the collection's long test text. They are inert document fixtures, not
malware samples.

| Local file | Lhasa path | SHA-256 |
| --- | --- | --- |
| `lh1-lharc113.lzh` | `test/archives/lharc113/lh1.lzh` | `594af58cb27d6a80e9bebd108cc82a8ab881a4b5549e4fc10379fbf7a0d83069` |
| `lh1-os9-h2.lzh` | `test/archives/lha_os9_211c/h2_lh1.lzh` | `e6aaa8a5e7520d8cda49c29befe8edf6570f2aac425154e23483582b086550af` |
| `lh1-lharc113-long.lzh` | `test/archives/lharc113/long.lzh` | `fe1d64ba8b8e67a32d3d2bfa910fc13a70d220d4bfe4f68987a6d9c3a77ebd03` |
| `lh1-morphos-h0.lzh` | `test/archives/morphos_lha_2717/h0_lh1.lzh` | `36503045e565c50a1f4cef66dd8c544fcb2ce4ac18f9cd4d55f4f0051b3d5b6e` |
| `lh1-morphos-h1.lzh` | `test/archives/morphos_lha_2717/h1_lh1.lzh` | `921399b4f348489d9ca1e5c0b93620164bbe6493e7eae48cee2ebaf71558ade0` |
| `lh1-morphos-h2.lzh` | `test/archives/morphos_lha_2717/h2_lh1.lzh` | `fba263ff370eee30485d4be73d13fbb9ceeb0edf5691253288e83884ab30dd21` |
| `lh1-amiga.lzh` | `test/archives/lha_amiga_212/lh1.lzh` | `4981a27f0e051b33d7813a630f970d820703f61c14f896dc5ed2a84f8240a836` |
| `lh1-osk-h2.lzh` | `test/archives/lha_osk_201/h2_lh1.lzh` | `82020122cf7ad688b4182fe26435c808c5f1b729eb16859dfdf42a173da36a81` |

The child SHA-256 values are pinned by the Go tests. Independent
ClamAV 1.5.4 whole-child HDB detections are in the
[v141 reference receipt](../../../../docs/edr/native-engine/audits/2026-10-28-v141-lh1/reference-verdicts.json).
