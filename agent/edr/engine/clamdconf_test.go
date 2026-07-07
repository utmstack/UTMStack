package engine

import (
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestRenderClamdConfHasAllDetectionEngines(t *testing.T) {
	cfg := config.Default() // ClamdAddr 127.0.0.1:3310
	tn := DeriveTuning(8*gib, 4, bigTemp, "")
	conf := RenderClamdConf(cfg, tn)

	// Every container-decomposition + engine directive must be present and ON (§3).
	must := []string{
		"Bytecode yes", "BytecodeSecurity TrustSigned",
		"ScanPE yes", "ScanELF yes", "ScanOLE2 yes", "ScanPDF yes",
		"ScanHTML yes", "ScanSWF yes", "ScanArchive yes", "ScanMail yes",
		"HeuristicAlerts yes", "PhishingSignatures yes", "PhishingScanURLs yes",
		"AlertExceedsMax yes", "AlertBrokenExecutables yes", "AlertEncryptedArchive yes",
	}
	for _, m := range must {
		if !strings.Contains(conf, m) {
			t.Fatalf("clamd.conf missing directive %q\n---\n%s", m, conf)
		}
	}
}

func TestRenderClamdConfSecurityBoundsStatic(t *testing.T) {
	cfg := config.Default()
	// Even a huge host must keep MaxRecursion 16 / MaxFiles 10000 (§6/§8).
	tn := DeriveTuning(128*gib, 32, bigTemp, "")
	conf := RenderClamdConf(cfg, tn)
	if !strings.Contains(conf, "MaxRecursion 16") {
		t.Fatalf("MaxRecursion must be static 16:\n%s", conf)
	}
	if !strings.Contains(conf, "MaxFiles 10000") {
		t.Fatalf("MaxFiles must be static 10000:\n%s", conf)
	}
}

func TestRenderClamdConfTransportAndDerivedKnobs(t *testing.T) {
	cfg := config.Default()
	tn := DeriveTuning(8*gib, 4, bigTemp, "")
	conf := RenderClamdConf(cfg, tn)

	for _, m := range []string{"TCPSocket 3310", "TCPAddr 127.0.0.1", "StreamMaxLength", "MaxThreads 4", "ConcurrentDatabaseReload yes"} {
		if !strings.Contains(conf, m) {
			t.Fatalf("clamd.conf missing %q\n---\n%s", m, conf)
		}
	}
}

func TestRenderClamdConfLowRAMDisablesConcurrentReload(t *testing.T) {
	cfg := config.Default()
	tn := DeriveTuning(2*gib, 2, bigTemp, "")
	conf := RenderClamdConf(cfg, tn)
	if !strings.Contains(conf, "ConcurrentDatabaseReload no") {
		t.Fatalf("low-RAM host must disable concurrent reload:\n%s", conf)
	}
}

func TestRenderFreshclamConfDefaultsToOfficialCDN(t *testing.T) {
	conf := RenderFreshclamConf(config.Default()) // no SigMirror
	for _, m := range []string{"ScriptedUpdates yes", "Checks 24", "DatabaseMirror database.clamav.net"} {
		if !strings.Contains(conf, m) {
			t.Fatalf("freshclam.conf missing %q\n---\n%s", m, conf)
		}
	}
	if strings.Contains(conf, "PrivateMirror") {
		t.Fatalf("default config must not use a private mirror:\n%s", conf)
	}
}

func TestRenderFreshclamConfUsesPrivateMirrorWhenSet(t *testing.T) {
	cfg := config.Default()
	cfg.SigMirror = "https://utmstack.example.com/clamav"
	conf := RenderFreshclamConf(cfg)
	if !strings.Contains(conf, "PrivateMirror https://utmstack.example.com/clamav") {
		t.Fatalf("expected private mirror directive:\n%s", conf)
	}
	if strings.Contains(conf, "DatabaseMirror database.clamav.net") {
		t.Fatalf("private mirror set → must not also use the official CDN:\n%s", conf)
	}
}
