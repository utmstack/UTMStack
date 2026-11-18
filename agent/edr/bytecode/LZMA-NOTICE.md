# Go LZMA streaming decoder

The streaming decoder adapts the LZMA wire algorithm and streaming behavior of
Igor Pavlov's public-domain LZMA SDK into Go. The native compatibility reference
is the 2009 SDK retained by ClamAV 1.5.4. Its original files identify the SDK as
public domain; see [the upstream SDK notice](https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/COPYING/COPYING.lzma).

The Go runtime does not link ClamAV or a native LZMA library. ClamAV's interface
is a behavioral reference; its test-only reference build, source identities and
original notices are retained with the corresponding audit evidence. The
existing `github.com/ulikunitz/xz/lzma` writer is also used in tests to generate
owned inputs, independently decoded by the pinned native reference.
