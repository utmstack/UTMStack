# PE resource lookup reference

The Go resource-directory walker uses the PE resource layout and the observed
lookup behavior of ClamAV 1.5.4 as references. It preserves signed selectors,
16-bit name/language counts, callback ordering, physical read boundaries and
the first matching type directory. Icon evaluation uses the enclosing Go scan
budget and stops after a proven match.

Reference: https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/libclamav/pe.c
SHA-256: `b5f2d8ba2fcc0eb08b887d89e6a7cdbbe84dcb973e3022217c85d77566399c26`.
The reference is copyright Cisco Systems and Sourcefire and carries GNU GPL
version 2 notices. Its source, copyright notices and accompanying license are
retained in the v266 audit as reference material. The Go executable does not
compile, link or invoke that C source or the ClamAV scanner.

The icon parser/hash reference remains:
https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/libclamav/pe_icons.c
SHA-256: `1dae6bfdc0dc77625a95387e3d628b1a6b70f6fc3518ff86fe78cb7e31be77dd`.
The v266 audit pins 384 full-scanner verdicts over 192 inert generated PEs,
including overflow counts, partial entries and malformed tails. This evidence
does not qualify every PE header, icon bitmap or native mathematical profile.
