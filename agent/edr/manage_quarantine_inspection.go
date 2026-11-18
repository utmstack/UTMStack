package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func runQuarantineInspection(args []string, out, diagnostic io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine inspect <id>")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "inspection configuration:", err)
		return 1
	}
	report, inspectionErr := quarantine.InspectRecovery(cfg.QuarantineDir, args[0])
	if err := json.NewEncoder(out).Encode(report); err != nil {
		fmt.Fprintln(diagnostic, "inspection output:", err)
		return 1
	}
	if inspectionErr != nil {
		fmt.Fprintln(diagnostic, "inspection incomplete:", inspectionErr)
		return 1
	}
	return 0
}

func runQuarantineFreshInspection(args []string, out, diagnostic io.Writer) int {
	if len(args) != 5 || args[1] != "--hold" || args[3] != "--destination" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine inspect-fresh <id> --hold <decision-uuid> --destination <absolute-local-path>")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh inspection configuration:", err)
		return 1
	}
	report, err := quarantine.InspectFreshRestore(cfg.QuarantineDir, args[0], args[2], args[4])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh inspection incomplete:", err)
		return 1
	}
	if err = json.NewEncoder(out).Encode(report); err != nil {
		fmt.Fprintln(diagnostic, "fresh inspection output:", err)
		return 1
	}
	return 0
}
