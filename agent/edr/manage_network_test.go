package main

import (
	"bytes"
	"context"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
	"path/filepath"
	"testing"
	"time"
)

type fakeIsolation struct {
	policy netblock.IsolationPolicy
	active bool
	calls  int
}

func (f *fakeIsolation) Isolate(_ context.Context, p netblock.IsolationPolicy) (netblock.IsolationStatus, error) {
	f.policy = p
	f.active = true
	f.calls++
	return f.Status(context.Background())
}
func (f *fakeIsolation) Release(context.Context) error { f.active = false; f.calls++; return nil }
func (f *fakeIsolation) Status(context.Context) (netblock.IsolationStatus, error) {
	return netblock.IsolationStatus{Supported: true, Active: f.active, Verified: true}, nil
}
func TestIsolationCLIRequiresFiniteExplicitReachability(t *testing.T) {
	oldFactory, oldConfig := newIsolationController, config.ConfigFile
	defer func() { newIsolationController = oldFactory; config.ConfigFile = oldConfig }()
	config.ConfigFile = filepath.Join(t.TempDir(), "missing-config.json")
	f := &fakeIsolation{}
	constructed := 0
	newIsolationController = func() (isolationController, error) { constructed++; return f, nil }
	for _, args := range [][]string{{"isolate"}, {"isolate", "--management", "bad"}, {"isolate", "--management", "192.0.2.1", "--duration", "2h"}, {"isolate", "--management", "192.0.2.1", "--allow", "0.0.0.0/0"}, {"release", "extra"}} {
		if runNetwork(args, &bytes.Buffer{}) == 0 {
			t.Fatal("invalid request accepted", args)
		}
	}
	if constructed != 0 {
		t.Fatal("invalid input invoked native backend")
	}
	if code := runNetwork([]string{"isolate", "--management", "192.0.2.1", "--duration", "1m"}, &bytes.Buffer{}); code != 0 {
		t.Fatal(code)
	}
	if !f.active || f.policy.Duration != time.Minute || len(f.policy.Management) != 1 {
		t.Fatalf("policy %+v", f.policy)
	}
	if runNetwork([]string{"release"}, &bytes.Buffer{}) != 0 || f.active {
		t.Fatal("release failed")
	}
}
