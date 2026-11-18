package netblock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

type failingArtifactReader struct{}

func (failingArtifactReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (failingArtifactReader) Close() error             { return nil }

func TestNetworkFeedPartialArtifactFailurePreservesWholeCycle(t *testing.T) {
	for _, daily := range []bool{false, true} {
		kind := "snapshot"
		if daily {
			kind = "daily"
		}
		for _, failure := range []string{"fetch", "read", "checksum-fetch", "checksum-mismatch"} {
			for _, failedLevel := range []string{"level1", "level2"} {
				t.Run(kind+"/"+failure+"/"+failedLevel, func(t *testing.T) {
					f, c, calls := newBoundedFeedTest(t)
					state, err := c.GetFeedState("level1/ip")
					if err != nil {
						t.Fatal(err)
					}
					previousSync := f.lastFullSync
					f.cfg.Blocklist.Levels = []int{1, 2}
					payload := gzList(t, "192.0.2.1\n")
					if daily {
						payload = []byte("{\"value\":\"198.51.100.4\",\"type\":\"ip\",\"op\":\"del\"}\n{\"value\":\"192.0.2.1\",\"type\":\"ip\",\"op\":\"add\"}\n")
					}
					broken := true
					f.Fetch = func(url string) (io.ReadCloser, error) {
						if broken && strings.Contains(url, "/"+failedLevel+"/") {
							switch failure {
							case "fetch":
								return nil, errors.New("owned fixture unavailable")
							case "read":
								return failingArtifactReader{}, nil
							}
						}
						return io.NopCloser(bytes.NewReader(payload)), nil
					}
					f.FetchChecksum = func(url string) (string, error) {
						if broken && strings.Contains(url, "/"+failedLevel+"/") {
							switch failure {
							case "checksum-fetch":
								return "", errors.New("owned checksum unavailable")
							case "checksum-mismatch":
								return strings.Repeat("0", 64), nil
							}
						}
						return sha(payload), nil
					}
					update := f.UpdateOnce
					if daily {
						update = f.updateDeltaOnce
					}
					if err := update(context.Background()); err == nil {
						t.Fatal("partial artifact cycle reported success")
					}
					assertBoundedFeedLastGood(t, f, c, *calls, state.Checksum)
					if !f.lastFullSync.Equal(previousSync) {
						t.Fatal("partial artifact cycle advanced feed freshness")
					}
					// Once every configured artifact is valid, publication resumes
					// and verified removals may replace stale enforcement.
					broken = false
					if err := update(context.Background()); err != nil {
						t.Fatal(err)
					}
					rows, err := c.ListIndicators()
					if err != nil || len(rows) != 1 || rows[0].Value != "192.0.2.1" || *calls != 1 || f.HealthError() != nil {
						t.Fatalf("recovered full cycle: rows=%+v error=%v apply=%d health=%v", rows, err, *calls, f.HealthError())
					}
				})
			}
		}
	}
}

func TestNetworkFeedMalformedRecordsCannotPublishVerifiedSubset(t *testing.T) {
	for _, daily := range []bool{false, true} {
		name := "snapshot"
		if daily {
			name = "daily"
		}
		t.Run(name, func(t *testing.T) {
			f, c, calls := newBoundedFeedTest(t)
			state, err := c.GetFeedState("level1/ip")
			if err != nil {
				t.Fatal(err)
			}
			payload := gzList(t, "192.0.2.1\nnot-an-address\n")
			if daily {
				payload = []byte("{\"value\":\"198.51.100.4\",\"op\":\"del\"}\n{broken\n")
			}
			f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
			f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
			update := f.UpdateOnce
			if daily {
				update = f.updateDeltaOnce
			}
			if err := update(context.Background()); err == nil {
				t.Fatal("a valid checksum must not turn malformed rows into a complete update")
			}
			assertBoundedFeedLastGood(t, f, c, *calls, state.Checksum)
		})
	}
}

func TestNetworkFeedInvalidNamesPreserveLastGood(t *testing.T) {
	for _, typ := range []string{"domain", "hostname"} {
		for _, daily := range []bool{false, true} {
			for _, name := range []string{".", "bad..example", "bad name.example", "bad/name.example", "*.example", "-bad.example", "bad-.example", "bad\x00.example", "é.example", strings.Repeat("a", 64) + ".example", strings.Repeat("abcd.", 51) + "example"} {
				t.Run(typ+"/"+name+fmt.Sprint(daily), func(t *testing.T) {
					f, c, calls := newBoundedFeedTest(t)
					state, err := c.GetFeedState("level1/ip")
					if err != nil {
						t.Fatal(err)
					}
					f.cfg.Blocklist.IndicatorTypes = []string{typ}
					payload := gzList(t, name+"\n")
					if daily {
						first, _ := json.Marshal(DailyOp{Value: "198.51.100.4", Type: "ip", Op: "del"})
						bad, _ := json.Marshal(DailyOp{Value: name, Type: typ, Op: "add"})
						payload = append(append(first, '\n'), bad...)
					}
					f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }
					f.FetchChecksum = func(string) (string, error) { return sha(payload), nil }
					update := f.UpdateOnce
					if daily {
						update = f.updateDeltaOnce
					}
					if err := update(context.Background()); err == nil {
						t.Fatal("invalid name published a fresh policy")
					}
					assertBoundedFeedLastGood(t, f, c, *calls, state.Checksum)
				})
			}
		}
	}
}

func TestNetworkFeedLiteralNamePolicyKeepsSupportedNames(t *testing.T) {
	for _, name := range []string{"GOOD.Example.", "host_1", "3host.example", "_ldap._tcp.example.test", "xn--bcher-kva.example", strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)} {
		for _, typ := range []string{"domain", "hostname"} {
			inds, _, err := parseAccumulativeBounded(gzListReader(t, name), 1, typ, 4096, 2)
			if err != nil || len(inds) != 1 || inds[0].Value != canonName(name) {
				t.Fatalf("supported snapshot %q: %+v %v", name, inds, err)
			}
			b, _ := json.Marshal(DailyOp{Value: name, Type: typ, Op: "add"})
			ops, _, err := parseDailyBounded(bytes.NewReader(b), 4096, 2)
			if err != nil || len(ops) != 1 {
				t.Fatalf("supported daily %q: %+v %v", name, ops, err)
			}
		}
	}
}

func gzListReader(t *testing.T, name string) io.Reader {
	t.Helper()
	return bytes.NewReader(gzList(t, name+"\n"))
}
