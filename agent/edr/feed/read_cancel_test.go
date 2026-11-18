package feed

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

type canceledEOFTransport func(*http.Request) (*http.Response, error)

func (f canceledEOFTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Some transports can finish a canceled stream with EOF rather than returning
// a cancellation error. A successful ReadAll must not erase the known cause.
type canceledEOFBody struct {
	ctx  context.Context
	sent bool
	data []byte
}

func (b *canceledEOFBody) Read(p []byte) (int, error) {
	if !b.sent {
		b.sent = true
		return copy(p, b.data), nil
	}
	<-b.ctx.Done()
	return 0, io.EOF
}
func (*canceledEOFBody) Close() error { return nil }

func TestReadRetainsCancellationCauseAfterEOF(t *testing.T) {
	for _, mode := range []string{"stall", "parent"} {
		t.Run(mode, func(t *testing.T) {
			parent, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			f := NewNative(config.Default(), nil, &testGenerationStore{})
			f.stallTimeout = 20 * time.Millisecond
			want := errTransferStalled
			f.client = &http.Client{Transport: canceledEOFTransport(func(r *http.Request) (*http.Response, error) {
				if mode == "parent" {
					want = context.Canceled
					cancel()
				}
				return &http.Response{StatusCode: http.StatusOK, ContentLength: -1,
					Header: make(http.Header), Body: &canceledEOFBody{ctx: r.Context(), data: []byte("partial fixture")}, Request: r}, nil
			})}
			got, err := f.read(parent, "https://owned-fixture.invalid/resource", 64)
			if !errors.Is(err, want) || len(got) != 0 {
				t.Fatalf("canceled transfer lost cause: bytes=%q err=%v want=%v", got, err, want)
			}
		})
	}
}

func TestStalledEOFDoesNotActivateCompleteArtifact(t *testing.T) {
	fix := newFeedFixture(t)
	fix.f.stallTimeout = 100 * time.Millisecond
	artifactReads := 0
	fix.f.client.Transport = canceledEOFTransport(func(r *http.Request) (*http.Response, error) {
		data, ok := fix.resources[r.URL.Path]
		if !ok {
			t.Fatalf("unexpected fixture request: %s", r.URL.Path)
		}
		var body io.ReadCloser = io.NopCloser(bytes.NewReader(data))
		if r.URL.Path == "/signatures/test.ndb" {
			artifactReads++
			// The complete artifact has the correct signed checksum. The
			// stream still stalls before EOF, so accepting its bytes would
			// incorrectly activate after this request's deadline.
			body = &canceledEOFBody{ctx: r.Context(), data: data}
		}
		return &http.Response{StatusCode: http.StatusOK, ContentLength: -1,
			Header: make(http.Header), Body: body, Request: r}, nil
	})
	beforeSequence, beforeHash := fix.store.FeedState()
	_, err := fix.f.UpdateOnce(context.Background())
	if artifactReads != 1 {
		t.Fatal("did not reach the complete artifact's stalled stream")
	}
	if !errors.Is(err, errTransferStalled) || fix.store.calls != 0 || !fix.f.LastSuccess().IsZero() {
		t.Fatalf("stalled EOF activated or changed success health: err=%v activations=%d success=%v", err, fix.store.calls, fix.f.LastSuccess())
	}
	if sequence, hash := fix.store.FeedState(); sequence != beforeSequence || hash != beforeHash || fix.store.Version() != "previous" {
		t.Fatal("stalled transfer changed the last-good generation")
	}
}

func TestReadRetainsRequestCancellationCause(t *testing.T) {
	for _, mode := range []string{"stall", "parent"} {
		t.Run(mode, func(t *testing.T) {
			parent, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			f := NewNative(config.Default(), nil, &testGenerationStore{})
			f.stallTimeout = 20 * time.Millisecond
			want := errTransferStalled
			f.client = &http.Client{Transport: canceledEOFTransport(func(r *http.Request) (*http.Response, error) {
				if mode == "parent" {
					want = context.Canceled
					cancel()
				}
				<-r.Context().Done()
				return nil, io.EOF
			})}
			got, err := f.read(parent, "https://owned-fixture.invalid/resource", 64)
			if !errors.Is(err, want) || len(got) != 0 {
				t.Fatalf("request lost cancellation cause: bytes=%q err=%v want=%v", got, err, want)
			}
		})
	}
}
