package bytecode

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBzip2MemberNativePublication(t *testing.T) {
	f, err := os.Open("testdata/bzip2_member_reference.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	archive := tar.NewReader(z)
	inputs := map[string][]byte{}
	var expected map[string]struct {
		Consumed int    `json:"consumed"`
		Size     int    `json:"output_bytes"`
		SHA256   string `json:"sha256"`
		Status   int    `json:"status"`
	}
	for {
		h, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil || !h.FileInfo().Mode().IsRegular() || h.Size > 8<<20 {
			t.Fatal("invalid native member archive", err)
		}
		data, err := io.ReadAll(archive)
		if err != nil {
			t.Fatal(err)
		}
		if h.Name == "expected.json" {
			if err := json.Unmarshal(data, &expected); err != nil {
				t.Fatal(err)
			}
		} else {
			name, ok := strings.CutPrefix(h.Name, "corpus/")
			if !ok || name == "" || inputs[name] != nil {
				t.Fatal("unexpected or duplicate native input")
			}
			inputs[name] = data
		}
	}
	if len(inputs) != 265 || len(expected) != 265 {
		t.Fatal("wrong native reference inventory")
	}
	for name, want := range expected {
		for _, window := range []int{7, 4096, 65536} {
			t.Run(fmt.Sprintf("%s/%d", name, window), func(t *testing.T) {
				r := NewBzip2MemberReader(context.Background(), inputs[name])
				defer r.Close()
				var output bytes.Buffer
				buffer := make([]byte, window)
				var terminal error
				for {
					n, err := r.Read(buffer)
					output.Write(buffer[:n])
					if err != nil {
						terminal = err
						break
					}
				}
				if r.CompressedBytes() != want.Consumed || output.Len() != want.Size ||
					fmt.Sprintf("%x", sha256.Sum256(output.Bytes())) != want.SHA256 ||
					(terminal == io.EOF) != (want.Status == bzipEnd) {
					t.Fatalf("native member differs: consumed=%d size=%d error=%v", r.CompressedBytes(), output.Len(), terminal)
				}
				if r.DecodedBytes() < r.DeliveredBytes() || r.DeliveredBytes() != int64(output.Len()) ||
					r.DecodedBytes()-r.DeliveredBytes() > 8192 || r.decoder.tt != nil {
					t.Fatal("decoder accounting or terminal cleanup failed")
				}
				before := r.CompressedBytes()
				for range 3 {
					if n, err := r.Read(buffer); n != 0 || err == nil || r.CompressedBytes() != before {
						t.Fatal("terminal reader reentered the decoder")
					}
				}
			})
		}
	}
}

type bzip2CancelWorkContext struct{ calls, after int }

func (c *bzip2CancelWorkContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *bzip2CancelWorkContext) Done() <-chan struct{}       { return nil }
func (c *bzip2CancelWorkContext) Value(any) any               { return nil }
func (c *bzip2CancelWorkContext) Err() error {
	c.calls++
	if c.calls >= c.after {
		return context.Canceled
	}
	return nil
}

func TestBzip2MemberCancellationDuringWork(t *testing.T) {
	// A valid harmless member retained from the independently encoded pipe
	// controls reaches the decoder's work checkpoints before publishing data.
	ctx := &bzip2CancelWorkContext{after: 3}
	r := NewBzip2MemberReader(ctx, bzip2TestPacked(t))
	defer r.Close()
	_, err := io.ReadAll(r)
	if !errors.Is(err, context.Canceled) || r.CompressedBytes() == 0 || r.decoder.tt != nil {
		t.Fatalf("work cancellation: error=%v consumed=%d retained=%d", err, r.CompressedBytes(), len(r.decoder.tt))
	}
}

func FuzzBzip2MemberReader(f *testing.F) {
	f.Add([]byte("BZh9"), uint16(1))
	packed, err := hex.DecodeString(bzip2TestHex)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(packed, uint16(511))
	f.Fuzz(func(t *testing.T, input []byte, size uint16) {
		if len(input) > 4096 {
			t.Skip()
		}
		r := NewBzip2MemberReader(context.Background(), input)
		defer r.Close()
		limit := int64(size%8192) + 1
		out, _ := io.ReadAll(io.LimitReader(r, limit))
		if r.CompressedBytes() < 0 || r.CompressedBytes() > len(input) || r.DeliveredBytes() != int64(len(out)) ||
			r.DecodedBytes() < r.DeliveredBytes() || r.DecodedBytes()-r.DeliveredBytes() > 8192 || int64(len(out)) > limit {
			t.Fatal("member bounds/accounting invariant failed")
		}
	})
}

func TestBzip2MemberCancellationAndClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewBzip2MemberReader(ctx, []byte("BZh9..."))
	if n, err := r.Read(make([]byte, 1)); n != 0 || !errors.Is(err, context.Canceled) || r.CompressedBytes() != 0 {
		t.Fatalf("cancelled reader: n=%d err=%v", n, err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if n, err := r.Read(make([]byte, 1)); n != 0 || !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("closed reader: n=%d err=%v", n, err)
	}
	if n, err := r.Read(nil); n != 0 || err != nil {
		t.Fatalf("zero-size read: n=%d err=%v", n, err)
	}
}
