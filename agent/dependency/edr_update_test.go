package dependency

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
)

func putEDR(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0700); err != nil {
		t.Fatal(err)
	}
}

func TestNativeEDRVersionGate(t *testing.T) {
	good := buildinfo.Info{PackageVersion: EDRVersion, Implementation: buildinfo.ImplementationVersion, OS: "windows", Architecture: "arm64", GoVersion: "go1.25.5"}
	current, err := strconv.Atoi(strings.TrimPrefix(buildinfo.ImplementationVersion, "native-go-v"))
	if err != nil || current < 1 {
		t.Fatalf("invalid current implementation version: %q", buildinfo.ImplementationVersion)
	}
	for _, field := range []string{"valid", "old package", "stale native implementation", "future native implementation", "external implementation", "wrong OS", "wrong architecture", "invalid JSON"} {
		t.Run(field, func(t *testing.T) {
			info := good
			switch field {
			case "old package":
				info.PackageVersion = "12.0.0"
			case "stale native implementation":
				info.Implementation = "native-go-v183"
			case "future native implementation":
				info.Implementation = "native-go-v" + strconv.Itoa(current+1)
			case "external implementation":
				info.Implementation = "external-daemon"
			case "wrong OS":
				info.OS = "linux"
			case "wrong architecture":
				info.Architecture = "amd64"
			}
			b, _ := json.Marshal(info)
			if field == "invalid JSON" {
				b = []byte("not JSON")
			}
			err := validateNativeEDRVersion(b, "windows", "arm64")
			if (err == nil) != (field == "valid") {
				t.Fatalf("gate returned %v", err)
			}
		})
	}
	if EDRVersion == "12.0.0" {
		t.Fatal("migration version cannot equal retired distribution")
	}
}

func TestNativeEDRDownloadIsStaged(t *testing.T) {
	d := nativeEDRDependency(t.TempDir())
	if d.Name != "edr" || d.Version != EDRVersion || d.DownloadName == filepath.Base(d.BinaryPath) || d.Download == nil || d.DownloadURL != nil || d.Recover == nil || d.Update == nil || d.PostDownload == nil || d.CheckExisting == nil {
		t.Fatalf("incomplete native migration dependency: %+v", d)
	}
}

func TestTrackedNativeEDRRechecksExactBuild(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "utmstack_edr")
	putEDR(t, path, "old implementation")
	d := Dependency{
		Name: "edr", Version: EDRVersion, BinaryPath: path,
		CheckExisting: func() bool {
			b, err := os.ReadFile(path)
			return err == nil && string(b) == "current implementation"
		},
	}
	tracked := &InstalledDep{Name: "edr", Version: EDRVersion}
	if !d.needsUpdate(tracked) {
		t.Fatal("tracked stale implementation was treated as current")
	}
	putEDR(t, path, "current implementation")
	if d.needsUpdate(tracked) {
		t.Fatal("current implementation would be downloaded again")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if !d.needsUpdate(tracked) {
		t.Fatal("tracked missing executable was treated as current")
	}
	putEDR(t, path, "current implementation")
	tracked.Version = "older-package"
	if !d.needsUpdate(tracked) {
		t.Fatal("older package version was treated as current")
	}
}

func TestNativeEDRVersionOutputBound(t *testing.T) {
	var b boundedVersionOutput
	if _, err := b.Write(make([]byte, 64<<10)); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write([]byte("x")); err == nil {
		t.Fatal("unbounded version output")
	}
	var copied boundedVersionOutput
	if _, err := io.Copy(&copied, strings.NewReader(strings.Repeat("x", 1<<20))); err == nil || len(copied.Bytes()) > 64<<10 {
		t.Fatal("copy bypassed the version output bound")
	}
}

func TestNativeEDRRejectsNonExecutableMetadataBeforeLaunch(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"missing", "empty", "directory", "oversized"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(base, name)
			switch name {
			case "empty":
				putEDR(t, path, "")
			case "directory":
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				f, err := os.Create(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.Truncate((256 << 20) + 1); err != nil {
					f.Close()
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err := verifyNativeEDR(path); err == nil {
				t.Fatal("invalid executable metadata accepted")
			}
		})
	}
}
