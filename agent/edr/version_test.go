package main

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func TestVersionIsInspectableWithoutDatabase(t *testing.T) {
	var out bytes.Buffer
	if err := runVersion([]string{"--json"}, &out); err != nil {
		t.Fatal(err)
	}
	var info buildinfo.Info
	if err := json.Unmarshal(out.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.PackageVersion != buildinfo.Version || info.Implementation != engine.ImplementationVersion || info.OS != runtime.GOOS || info.Architecture != runtime.GOARCH || info.GoVersion == "" {
		t.Fatalf("incomplete build identity: %+v", info)
	}
	out.Reset()
	if err := runVersion(nil, &out); err != nil || !strings.Contains(out.String(), buildinfo.Version) {
		t.Fatalf("text %q, %v", out.String(), err)
	}
	out.Reset()
	if err := runVersion([]string{"--bad"}, &out); err == nil || out.Len() != 0 {
		t.Fatalf("bad option accepted: %q, %v", out.String(), err)
	}
}
