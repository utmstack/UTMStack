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

func runQuarantineHold(args []string, out, diagnostic io.Writer) int {
	// Keep this deliberately explicit: no default case file, inferred identity,
	// or hash-only approval can authorize a retention decision.
	if len(args) != 5 || args[1] != "--inspection" || args[3] != "--reason" {
		fmt.Fprintln(diagnostic, "usage: utmstack_edr quarantine hold <id> --inspection <saved-json> --reason <text>")
		return 1
	}
	f, err := os.Open(args[2])
	if err != nil {
		fmt.Fprintln(diagnostic, "review input:", err)
		return 1
	}
	// Three bounded journals, JSON base64 expansion and the larger hold fence.
	const maxCase = 8 << 20
	info, statErr := f.Stat()
	if statErr != nil || !info.Mode().IsRegular() || info.Size() > maxCase {
		f.Close()
		fmt.Fprintln(diagnostic, "review input must be a regular JSON file within eight MiB")
		return 1
	}
	raw, readErr := io.ReadAll(io.LimitReader(f, maxCase+1))
	closeErr := f.Close()
	if readErr != nil || closeErr != nil || len(raw) > maxCase {
		fmt.Fprintln(diagnostic, "review input:", errors.Join(readErr, closeErr, errors.New("incomplete or oversized case file")))
		return 1
	}
	var report quarantine.RecoveryInspection
	if err := json.Unmarshal(raw, &report); err != nil || report.Version != 1 || report.QuarantineID != args[0] {
		fmt.Fprintln(diagnostic, "review input must be a version-1 inspection for this exact quarantine ID")
		return 1
	}
	var approval quarantine.RecoveryApproval
	count := 0
	for _, entry := range report.Journals {
		if entry.Operation == "restore" {
			count++
			sum := sha256.Sum256(entry.Manifest)
			if entry.Status != "captured" || hex.EncodeToString(sum[:]) != entry.SHA256 {
				fmt.Fprintln(diagnostic, "review input has incomplete or changed restore evidence")
				return 1
			}
			approval = quarantine.RecoveryApproval{RootID: report.RootID, DirectoryID: entry.DirectoryID, ManifestID: entry.ManifestID, SHA256: entry.SHA256}
		}
	}
	if count != 1 {
		fmt.Fprintln(diagnostic, "review input requires exactly one captured restore journal")
		return 1
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(diagnostic, "review configuration:", err)
		return 1
	}
	decision, holdErr := quarantine.HoldRecovery(cfg.QuarantineDir, args[0], approval, args[4])
	if holdErr != nil {
		// A failed write can have retained a private preparation or published
		// fence. Preserve its decision ID for follow-up inspection.
		fmt.Fprintln(diagnostic, "retention decision:", holdErr, "decision_id="+decision.DecisionID)
		return 1
	}
	if err := json.NewEncoder(out).Encode(decision); err != nil {
		fmt.Fprintln(diagnostic, "decision saved but output failed:", err)
		return 1
	}
	return 0
}
