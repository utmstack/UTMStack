package engine

import "github.com/utmstack/UTMStack/agent/edr/bytecode"

// bytecodeEngineOptions reports only scan modules this native scanner enables.
// The masks use ClamAV's deprecated bytecode API layout, not this engine's
// internal parser flags. HWP3 and heuristic bits remain clear because those
// behaviors are not enabled as ClamAV-compatible modules here. A parser that
// reaches an unsupported member still reports incomplete analysis separately.
func (s *scanState) bytecodeEngineOptions() *bytecode.EngineOptions {
	const scan = 0x1 | // archive
		0x2 | // mail
		0x4 | // OLE2
		0x10 | // HTML
		0x20 | // PE
		0x2000 | // ELF
		0x4000 | // PDF
		0x400000 | // SWF
		0x1000000 // XML documents
	const database = 0x2 | // phishing
		0x8 | // phishing URLs
		0x2000 // bytecode
	options := uint32(scan)
	if s.collectMetadata && !s.metadataSuspended {
		// The deprecated bytecode API reports CL_SCAN_FILE_PROPERTIES when
		// CL_SCAN_GENERAL_COLLECT_METADATA is enabled for this scan.
		options |= 0x10000000
	}
	return &bytecode.EngineOptions{Scan: options, Database: database}
}
