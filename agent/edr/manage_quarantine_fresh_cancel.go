package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func runQuarantineFreshCancellationInspection(args []string, out, diagnostic io.Writer) int {
	if len(args) != 3 || args[1] != "--operation" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine inspect-fresh-cancel <id> --operation <new-operation-uuid>")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation inspection configuration:", err)
		return 1
	}
	result, err := quarantine.InspectFreshCancellation(cfg.QuarantineDir, args[0], args[2])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation inspection:", err)
		return 1
	}
	if err = json.NewEncoder(out).Encode(result); err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation inspection output:", err)
		return 1
	}
	return 0
}

func runQuarantineFreshCancellation(args []string, out, diagnostic io.Writer) int {
	if len(args) != 7 || args[1] != "--operation" || args[3] != "--inspection" || args[5] != "--reason" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine cancel-fresh <id> --operation <new-operation-uuid> --inspection <saved-cancellation-json> --reason <text>")
		return 1
	}
	body, err := readBoundedFreshReviewInput(args[4], 64<<20)
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation input:", err)
		return 1
	}
	var inspection quarantine.FreshCancellationInspection
	if err = json.Unmarshal(body, &inspection); err != nil || inspection.QuarantineID != args[0] || inspection.OperationID != args[2] {
		fmt.Fprintln(diagnostic, "fresh cancellation input must describe this exact quarantine ID and operation:", err)
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation configuration:", err)
		return 1
	}
	result, err := quarantine.CancelFreshRestore(cfg.QuarantineDir, args[0], args[2], inspection, args[6])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation:", err, "operation_id="+args[2])
		return 1
	}
	if err = json.NewEncoder(out).Encode(result); err != nil {
		fmt.Fprintln(diagnostic, "fresh cancellation recorded but output failed; preserve the exact inspection and retry:", err)
		return 1
	}
	return 0
}
