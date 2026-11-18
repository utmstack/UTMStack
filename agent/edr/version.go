package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func runVersion(args []string, out io.Writer) error {
	info := buildinfo.Current(engine.ImplementationVersion)
	if len(args) == 1 && args[0] == "--json" {
		return json.NewEncoder(out).Encode(info)
	}
	if len(args) != 0 {
		return fmt.Errorf("usage: utmstack_edr version [--json]")
	}
	_, err := fmt.Fprintf(out, "UTMStack EDR %s (%s, %s/%s)\n", info.PackageVersion, info.Implementation, info.OS, info.Architecture)
	return err
}
