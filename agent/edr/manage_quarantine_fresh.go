package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func readFreshReviewInput(path string) ([]byte, error) {
	return readBoundedFreshReviewInput(path, 24<<20)
}

func readBoundedFreshReviewInput(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.Join(err, errors.New("fresh review input must be regular JSON within its size limit"))
	}
	body, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(body)) > limit {
		return nil, errors.Join(err, errors.New("incomplete or oversized fresh review input"))
	}
	return body, nil
}

func runQuarantineFreshApproval(args []string, out, diagnostic io.Writer) int {
	if len(args) != 7 || args[1] != "--operation" || args[3] != "--inspection" || args[5] != "--reason" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine approve-fresh <id> --operation <new-uuid> --inspection <saved-fresh-json> --reason <text>")
		return 1
	}
	body, err := readFreshReviewInput(args[4])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh approval input:", err)
		return 1
	}
	var inspection quarantine.FreshRestoreInspection
	if err = json.Unmarshal(body, &inspection); err != nil || inspection.QuarantineID != args[0] {
		fmt.Fprintln(diagnostic, "fresh approval input must describe this exact quarantine ID:", err)
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh approval configuration:", err)
		return 1
	}
	a, err := quarantine.ApproveFreshRestore(cfg.QuarantineDir, args[0], args[2], inspection, args[6])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh approval:", err, "operation_id="+args[2])
		return 1
	}
	if err = json.NewEncoder(out).Encode(a); err != nil {
		fmt.Fprintln(diagnostic, "fresh approval saved but output failed:", err)
		return 1
	}
	return 0
}

func runQuarantineFreshExecution(args []string, out, diagnostic io.Writer) int {
	if len(args) != 5 || args[1] != "--approval" || args[3] != "--reason" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine execute-fresh <id> --approval <saved-approval-json> --reason <text>")
		return 1
	}
	body, err := readFreshReviewInput(args[2])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh execution input:", err)
		return 1
	}
	var a quarantine.FreshRestoreApproval
	if err = json.Unmarshal(body, &a); err != nil || a.QuarantineID != args[0] {
		fmt.Fprintln(diagnostic, "fresh execution input must describe this exact quarantine ID:", err)
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh execution configuration:", err)
		return 1
	}
	result, err := quarantine.ExecuteFreshRestore(cfg.QuarantineDir, config.DBFile, body, args[4])
	if err != nil {
		fmt.Fprintln(diagnostic, "fresh execution:", err, "operation_id="+a.OperationID, "phase="+result.Phase, "output_id="+result.OutputID)
		return 1
	}
	if err = json.NewEncoder(out).Encode(result); err != nil {
		fmt.Fprintln(diagnostic, "fresh execution completed but output failed; preserve the approval and retry exactly:", err)
		return 1
	}
	return 0
}
