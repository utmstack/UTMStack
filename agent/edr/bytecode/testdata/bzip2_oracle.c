/* Laboratory-only adapter to the independently installed libbz2.
 * No part of this file or native library is linked into the Go scanner. */
#include <bzlib.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {
    if (argc != 3) return 2;
    int iw = atoi(argv[1]), ow = atoi(argv[2]);
    if (iw < 1 || ow < 1 || iw > 1048576 || ow > 1048576) return 2;
    size_t size = 0, cap = 4096;
    unsigned char *input = malloc(cap), *output = malloc(ow);
    if (!input || !output) return 3;
    for (;;) {
        if (size == cap) {
            if (cap >= 16777216) return 4;
            cap *= 2;
            void *p = realloc(input, cap);
            if (!p) return 3;
            input = p;
        }
        size_t n = fread(input + size, 1, cap - size, stdin);
        size += n;
        if (!n) break;
    }
    bz_stream s;
    memset(&s, 0, sizeof(s));
    if (BZ2_bzDecompressInit(&s, 0, 0) != BZ_OK) return 5;
    size_t offset = 0;
    for (unsigned calls = 0; calls < 2000000; calls++) {
        unsigned available = (unsigned)(size - offset);
        if (available > (unsigned)iw) available = (unsigned)iw;
        s.next_in = (char *)input + offset;
        s.avail_in = available;
        s.next_out = (char *)output;
        s.avail_out = ow;
        /* ClamAV's adapter refuses empty input/output before calling libbz2. */
        int result = available ? BZ2_bzDecompress(&s) : -1;
        unsigned read = available - s.avail_in, written = ow - s.avail_out;
        printf("%u %u %d", read, written, result);
        for (unsigned i = 0; i < written; i++) printf(" %02x", output[i]);
        putchar('\n');
        offset += read;
        if (result != BZ_OK || (!read && !written)) break;
    }
    BZ2_bzDecompressEnd(&s);
    free(output);
    free(input);
    return 0;
}
