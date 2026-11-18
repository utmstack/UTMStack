/* Test-only zlib call-boundary oracle. The production scanner never links zlib. */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <zlib.h>

int main(int argc, char **argv) {
    if (argc != 4 && argc != 5) return 2;
    int bits = atoi(argv[1]);
    size_t in_chunk = (size_t)atoi(argv[2]);
    size_t out_chunk = (size_t)atoi(argv[3]);
    if (!in_chunk || !out_chunk || out_chunk > (1u << 20)) return 2;
    size_t len = 0, cap = 4096;
    unsigned char *input = malloc(cap), *output = malloc(out_chunk);
    if (!input || !output) return 3;
    for (;;) {
        if (len == cap) {
            cap *= 2;
            unsigned char *next = realloc(input, cap);
            if (!next) return 3;
            input = next;
        }
        size_t n = fread(input + len, 1, cap - len, stdin);
        len += n;
        if (n == 0) {
            if (ferror(stdin)) return 3;
            break;
        }
    }
    z_stream stream;
    memset(&stream, 0, sizeof(stream));
    if (inflateInit2(&stream, bits) != Z_OK) return 4;
    size_t pos = 0;
    int recover = argc == 5 && strcmp(argv[4], "sync") == 0;
    int need_sync = 0;
    for (int step = 0; step < 200000; step++) {
        size_t avail = len - pos;
        if (avail > in_chunk) avail = in_chunk;
        stream.next_in = input + pos;
        stream.avail_in = (uInt)avail;
        stream.next_out = output;
        stream.avail_out = (uInt)out_chunk;
        int status;
        if (recover) {
            for (;;) {
                if (!need_sync) {
                    status = inflate(&stream, Z_NO_FLUSH);
                    if (status == Z_DATA_ERROR) need_sync = 1;
                }
                if (need_sync) {
                    status = inflateSync(&stream);
                    if (status == Z_OK) {
                        need_sync = 0;
                        continue;
                    }
                }
                break;
            }
        } else {
            status = inflate(&stream, Z_NO_FLUSH);
        }
        size_t consumed = avail - stream.avail_in;
        size_t produced = out_chunk - stream.avail_out;
        pos += consumed;
        printf("%zu %zu %d", consumed, produced, status);
        for (size_t i = 0; i < produced; i++) printf(" %02x", output[i]);
        putchar('\n');
        if (status == Z_STREAM_END || (!recover && status < 0) ||
            (recover && status < 0 && pos == len) || (!consumed && !produced)) break;
    }
    inflateEnd(&stream);
    free(input);
    free(output);
    return 0;
}
