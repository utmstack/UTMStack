package service

import "os"

// defaultTrustedProcesses is a conservative built-in allowlist of processes that
// legitimately perform mass file operations or run recovery/backup commands, and
// would otherwise be common ransomware-guard false positives. Entries are matched
// by the trusted-process matcher (path subtree / base name / glob,
// case-insensitive, separator-agnostic), merged on top of cfg.Allowlist.Processes.
//
// Deliberately conservative for safety (anything under a trusted path is
// trusted): OS infrastructure is pinned to full System32 paths; third-party
// suites to their admin-writable Program Files subtrees — never bare base names
// that malware could impersonate. It intentionally does NOT trust vssadmin.exe /
// wbadmin.exe / bcdedit.exe themselves (those are the T1490 tells, and T1490 is
// scored against the *parent* encryptor, not the tool).
func defaultTrustedProcesses() []string {
	sr := os.Getenv("SystemRoot")
	if sr == "" {
		sr = `C:\Windows`
	}
	sys := sr + `\System32\`
	return []string{
		// Windows Volume Shadow Copy / backup / System Restore infrastructure —
		// creates shadow copies and touches many files during legitimate backups.
		sys + "vssvc.exe",
		sys + "wbengine.exe",
		sys + "swprv.exe",
		sys + "SrTasks.exe",
		// Windows Search indexer — mass file access.
		sys + "SearchIndexer.exe",
		sys + "SearchProtocolHost.exe",
		sys + "SearchFilterHost.exe",
		// Microsoft Defender (mass on-access scanning).
		`C:\Program Files\Windows Defender`,
		`C:\ProgramData\Microsoft\Windows Defender\Platform`,
		// Common third-party backup / imaging / disaster-recovery suites (subtree).
		`C:\Program Files\Veeam`,
		`C:\Program Files (x86)\Veeam`,
		`C:\Program Files\Acronis`,
		`C:\Program Files (x86)\Acronis`,
		`C:\Program Files\Macrium`,
		`C:\Program Files\Commvault`,
		`C:\Program Files\Veritas`,
		`C:\Program Files\Cohesity`,
		`C:\Program Files\Rubrik`,
		`C:\Program Files\Datto`,
		// File-sync clients (mass create/rename during sync).
		`C:\Program Files\Microsoft OneDrive`,
		`C:\Program Files\Dropbox`,
		`C:\Program Files (x86)\Dropbox`,
		`C:\Program Files\Google\Drive File Stream`,
		`C:\Program Files\Box\Box`,
	}
}
