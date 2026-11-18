package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func runQuarantineRetirementInspection(args []string, out, diagnostic io.Writer) int {
	if len(args) != 5 || args[1] != "--hold" || args[3] != "--stage" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine inspect-retirement <id> --hold <original-hold-uuid> --stage <retained-original-stage>")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "retirement inspection configuration:", err)
		return 1
	}
	result, err := quarantine.InspectCancelledRetirement(cfg.QuarantineDir, args[0], args[2], args[4])
	if err != nil {
		fmt.Fprintln(diagnostic, "retirement inspection:", err)
		return 1
	}
	if err = json.NewEncoder(out).Encode(result); err != nil {
		fmt.Fprintln(diagnostic, "retirement inspection output:", err)
		return 1
	}
	return 0
}

func runQuarantineRetirement(args []string, out, diagnostic io.Writer) int {
	if (len(args) != 5 && len(args) != 7) || args[1] != "--inspection" || args[3] != "--reason" || len(args) == 7 && args[5] != "--stage" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine retire-cancelled <id> --inspection <saved-retirement-json> --reason <text> [--stage <same-renamed-original-stage>]")
		return 1
	}
	body, err := readBoundedFreshReviewInput(args[2], 16<<20)
	if err != nil {
		fmt.Fprintln(diagnostic, "retirement input:", err)
		return 1
	}
	var inspection quarantine.CancelledRetirementInspection
	if err = json.Unmarshal(body, &inspection); err != nil || inspection.QuarantineID != args[0] {
		fmt.Fprintln(diagnostic, "retirement input must describe this exact quarantine ID:", err)
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "retirement configuration:", err)
		return 1
	}
	stage := ""
	if len(args) == 7 {
		stage = args[6]
	}
	result, err := quarantine.RetireCancelledRecovery(cfg.QuarantineDir, inspection, args[4], stage)
	if err != nil {
		fmt.Fprintln(diagnostic, "retirement incomplete; preserve the exact saved inspection and retained data:", err)
		return 1
	}
	if err = json.NewEncoder(out).Encode(result); err != nil {
		fmt.Fprintln(diagnostic, "retirement recorded but output failed; retry the exact saved inspection and reason:", err)
		return 1
	}
	return 0
}
