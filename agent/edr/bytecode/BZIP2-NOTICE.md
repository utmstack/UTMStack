# Bzip2 streaming decoder

The Go streaming decoder is an independent implementation of the bzip2 wire
format. Its legacy randomized-block constants are retained from the bzip2
1.0.8 `randtable.c` table, copyright Julian R Seward, under the accompanying
[bzip2 license](BZIP2-LICENSE.txt). This is an altered representation of that
data, not the original bzip2 source or library.

ClamAV 1.5.4's bytecode pipe adapter and libbz2 1.0.8 are behavior references.
The C file in testdata is a laboratory adapter to an independently installed
libbz2. Neither ClamAV nor libbz2 is linked into the Go scanner.
