package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

func runQuarantineRelease(args []string, out, diagnostic io.Writer) int {
	return runQuarantineReviewAction(args, out, diagnostic, false)
}

func runQuarantineCancel(args []string, out, diagnostic io.Writer) int {
	return runQuarantineReviewAction(args, out, diagnostic, true)
}

func runQuarantineReviewAction(args []string, out, diagnostic io.Writer, cancel bool) int {
	action := "release-hold"
	if cancel {
		action = "cancel-restore"
	}
	if len(args) != 7 || args[1] != "--hold" || args[3] != "--inspection" || args[5] != "--reason" {
		fmt.Fprintf(diagnostic, "usage: utmstack_edr quarantine %s <id> --hold <decision-uuid> --inspection <saved-json> --reason <text>\n", action)
		return 1
	}
	f, err := os.Open(args[4])
	if err != nil {
		fmt.Fprintln(diagnostic, action+" input:", err)
		return 1
	}
	const maxCase = 8 << 20
	info, statErr := f.Stat()
	if statErr != nil || !info.Mode().IsRegular() || info.Size() > maxCase {
		f.Close()
		fmt.Fprintln(diagnostic, action+" input must be a regular JSON file within eight MiB")
		return 1
	}
	raw, readErr := io.ReadAll(io.LimitReader(f, maxCase+1))
	closeErr := f.Close()
	if readErr != nil || closeErr != nil || len(raw) > maxCase {
		fmt.Fprintln(diagnostic, action+" input:", errors.Join(readErr, closeErr, errors.New("incomplete or oversized case file")))
		return 1
	}
	var report quarantine.RecoveryInspection
	if err := json.Unmarshal(raw, &report); err != nil || report.Version != 1 || report.QuarantineID != args[0] {
		fmt.Fprintln(diagnostic, action+" input must be a version-1 inspection for this exact quarantine ID")
		return 1
	}
	a := quarantine.RecoveryReleaseApproval{HoldDecisionID: args[2]}
	counts := map[string]int{}
	for _, entry := range report.Journals {
		if entry.Operation != "restore" && entry.Operation != "purge" {
			continue
		}
		counts[entry.Operation]++
		sum := sha256.Sum256(entry.Manifest)
		if entry.Status != "captured" || hex.EncodeToString(sum[:]) != entry.SHA256 {
			fmt.Fprintln(diagnostic, action+" input has incomplete or changed evidence")
			return 1
		}
		if entry.Operation == "restore" {
			a.Original = quarantine.RecoveryApproval{RootID: report.RootID, DirectoryID: entry.DirectoryID, ManifestID: entry.ManifestID, SHA256: entry.SHA256}
		} else {
			if entry.EntryKind != "file" {
				fmt.Fprintln(diagnostic, action+" input requires an exact regular retention fence")
				return 1
			}
			a.HoldDirectoryID, a.HoldFileID, a.HoldSHA256 = entry.DirectoryID, entry.ManifestID, entry.SHA256
		}
	}
	if counts["restore"] != 1 || counts["purge"] != 1 {
		fmt.Fprintln(diagnostic, action+" input requires exactly one captured restore journal and retention fence")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, action+" configuration:", err)
		return 1
	}
	var decision any
	decisionID := ""
	if cancel {
		var m quarantine.RecoveryCancellation
		m, err = quarantine.CancelRecovery(cfg.QuarantineDir, args[0], a, args[6])
		decision, decisionID = m, m.DecisionID
	} else {
		var m quarantine.RecoveryRelease
		m, err = quarantine.ReleaseRecovery(cfg.QuarantineDir, args[0], a, args[6])
		decision, decisionID = m, m.DecisionID
	}
	if err != nil {
		fmt.Fprintln(diagnostic, action+":", err, "hold_decision_id="+a.HoldDecisionID, "decision_id="+decisionID)
		return 1
	}
	if err := json.NewEncoder(out).Encode(decision); err != nil {
		fmt.Fprintln(diagnostic, action+" saved but output failed:", err)
		return 1
	}
	return 0
}
