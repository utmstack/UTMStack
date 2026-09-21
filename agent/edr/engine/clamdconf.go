package engine

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// engine on-disk layout under config.EngineDir
func databaseDir() string { return filepath.Join(config.EngineDir, "database") }
func tempDir() string     { return filepath.Join(config.EngineDir, "tmp") }

// RenderClamdConf produces a detection-maximizing, performance-safe clamd.conf
// (ClamAV Configuration & Tuning Guide §3/§5/§6/§7/§12). Feature and security
// directives are static; resource directives come from the derived Tuning.
func RenderClamdConf(cfg config.EDRConfig, t Tuning) string {
	host, port, err := net.SplitHostPort(cfg.ClamdAddr)
	if err != nil {
		host, port = "127.0.0.1", "3310"
	}
	streamMax := t.MaxScanSizeMB
	if streamMax < 100 {
		streamMax = 100 // headroom for AMSI/INSTREAM buffers (§12)
	}
	reload := "no"
	if t.ConcurrentReload {
		reload = "yes"
	}

	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	w("# UTMStack EDR engine configuration — auto-generated, do not edit by hand")
	w(fmt.Sprintf("# tier=%s resident=%v reason=%s", t.Tier, t.ResidentViable, t.Reason))
	w("")
	w("# --- transport (TCP on Windows) ---")
	w("TCPSocket " + port)
	w("TCPAddr " + host)
	w(fmt.Sprintf("StreamMaxLength %dM", streamMax))
	w("DatabaseDirectory " + databaseDir())
	w("TemporaryDirectory " + tempDir())
	w("LogTime yes")
	w("")
	w("# --- detection engines (static §3) ---")
	for _, d := range []string{
		"Bytecode yes",
		"BytecodeSecurity TrustSigned",
		"ScanPE yes",
		"ScanELF yes",
		"ScanOLE2 yes",
		"ScanPDF yes",
		"ScanHTML yes",
		"ScanSWF yes",
		"ScanXMLDOCS yes",
		"ScanArchive yes",
		"ScanMail yes",
		"HeuristicAlerts yes",
		"PhishingSignatures yes",
		"PhishingScanURLs yes",
	} {
		w(d)
	}
	w("")
	w("# --- heuristic alerts (selective §5; review-only, never auto-delete) ---")
	w("AlertExceedsMax yes")
	w("AlertBrokenExecutables yes")
	w("AlertEncryptedArchive yes")
	w("")
	w("# --- scan limits: security bounds static, sizes derived (§6/§8) ---")
	w(fmt.Sprintf("MaxScanSize %dM", t.MaxScanSizeMB))
	w(fmt.Sprintf("MaxFileSize %dM", t.MaxFileSizeMB))
	w(fmt.Sprintf("MaxRecursion %d", FixedMaxRecursion))
	w(fmt.Sprintf("MaxFiles %d", FixedMaxFiles))
	w(fmt.Sprintf("MaxScanTime %d", FixedMaxScanTime))
	w("")
	w("# --- performance / resources (derived §7/§8) ---")
	w(fmt.Sprintf("MaxThreads %d", t.MaxThreads))
	w(fmt.Sprintf("MaxQueue %d", t.MaxQueue))
	w("SelfCheck 3600")
	w("ConcurrentDatabaseReload " + reload)
	// DisableCache defaults to no (cache stays enabled — §7).
	return b.String()
}

// ResolveMirrorURL returns the effective private-mirror URL for signature
// updates: "" ⇒ auto-derive from the registered UTMStack server (choosing the
// transport by cfg.SkipCertValidate); "cdn" ⇒ none (use the official public
// database); anything else is explicit.
func ResolveMirrorURL(cfg config.EDRConfig) string {
	switch cfg.SigMirror {
	case "cdn":
		return ""
	case "":
		if cfg.Server == "" {
			return ""
		}
		if cfg.SkipCertValidate {
			// Self-signed deployment: freshclam cannot skip TLS verification, so
			// point it at the server's plain-HTTP signature port. The CVD's own
			// ClamAV signature guarantees integrity, not the transport.
			return "http://" + cfg.Server + ":" + config.MirrorHTTPPort + config.MirrorBasePath + "/signatures"
		}
		// Validated TLS: freshclam checks the server cert against the OS trust
		// store natively, like the agent's other channels.
		return "https://" + cfg.Server + ":" + config.MirrorPort + config.MirrorBasePath + "/signatures"
	default:
		return cfg.SigMirror
	}
}

// RenderFreshclamConf produces the signature-updater config (§9): current
// databases, incremental cdiff updates, and a sane (hourly) cadence.
func RenderFreshclamConf(cfg config.EDRConfig) string {
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	w("# UTMStack EDR signature-updater configuration — auto-generated")
	w("DatabaseDirectory " + databaseDir())
	if mirror := ResolveMirrorURL(cfg); mirror != "" {
		// Private mirror (the UTMStack server) — one upstream fetch fans out to
		// the fleet, avoids the official CDN's rate limits, and works on
		// restricted/air-gapped networks (§9).
		w("PrivateMirror " + mirror)
	} else {
		// Official public database (explicit "cdn", or no server registered).
		w("DatabaseMirror database.clamav.net")
	}
	w("ScriptedUpdates yes") // efficient incremental (cdiff) updates
	w("Checks 24")           // ~hourly; the CDN rate-limits, do not over-poll
	w("CompressLocalDatabase no")
	w("DatabaseOwner 0")
	return b.String()
}

// RenderFreshclamConfCDN forces the official public database regardless of
// mirror config — used for a single fallback cycle when the mirror is down.
func RenderFreshclamConfCDN(cfg config.EDRConfig) string {
	forced := cfg
	forced.SigMirror = "cdn"
	return RenderFreshclamConf(forced)
}

// WriteFreshclamConf writes freshclam.conf, optionally forcing the official
// public database for a fallback cycle while the private mirror is down.
func WriteFreshclamConf(cfg config.EDRConfig, useCDN bool) error {
	conf := RenderFreshclamConf(cfg)
	if useCDN {
		conf = RenderFreshclamConfCDN(cfg)
	}
	return os.WriteFile(filepath.Join(config.EngineDir, "freshclam.conf"), []byte(conf), 0o644)
}

// WriteEngineConfigs writes clamd.conf + freshclam.conf and ensures the
// database/temp directories exist. Called before starting the daemon.
func WriteEngineConfigs(cfg config.EDRConfig, t Tuning) error {
	for _, d := range []string{config.EngineDir, databaseDir(), tempDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("engine dir %s: %w", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(config.EngineDir, "clamd.conf"), []byte(RenderClamdConf(cfg, t)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(config.EngineDir, "freshclam.conf"), []byte(RenderFreshclamConf(cfg)), 0o644); err != nil {
		return err
	}
	return nil
}
