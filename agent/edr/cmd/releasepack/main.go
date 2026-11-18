// releasepack signs an already-built native EDR release. Production Windows
// inputs must be Authenticode-signed before this command is run.
package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	"github.com/utmstack/UTMStack/agent/edr/release"
)

func main() {
	var want release.Identity
	var source, output, keyPath string
	flag.StringVar(&want.PackageVersion, "version", buildinfo.Version, "exact distribution version")
	flag.StringVar(&want.Implementation, "implementation", buildinfo.ImplementationVersion, "exact native engine implementation")
	flag.StringVar(&want.OS, "os", "", "target OS")
	flag.StringVar(&want.Architecture, "arch", "", "target architecture")
	flag.StringVar(&want.ServiceName, "service", "", "service executable filename")
	flag.StringVar(&want.ProviderName, "provider", "", "AMSI DLL filename; required for Windows")
	flag.StringVar(&source, "source", "", "directory of already-signed release files")
	flag.StringVar(&output, "out", "", "new release archive path")
	flag.StringVar(&keyPath, "key", "", "file holding base64 Ed25519 private key")
	flag.Parse()
	if source == "" || output == "" || keyPath == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "releasepack requires -source, -out, -key and exact target identity")
		os.Exit(2)
	}
	encoded, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot read publisher signing key:", err)
		os.Exit(1)
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil || len(key) != ed25519.PrivateKeySize {
		fmt.Fprintln(os.Stderr, "publisher signing key must be a base64 Ed25519 private key")
		os.Exit(1)
	}
	manifest, err := release.Pack(output, source, ed25519.PrivateKey(key), want)
	if err != nil {
		fmt.Fprintln(os.Stderr, "release packaging failed:", err)
		os.Exit(1)
	}
	public := hex.EncodeToString(ed25519.PrivateKey(key).Public().(ed25519.PublicKey))
	for i := range key {
		key[i] = 0
	}
	fmt.Printf("signed %s/%s %s %s with %d artifacts; publisher public key %s\n", manifest.OS, manifest.Architecture, manifest.PackageVersion, manifest.Implementation, len(manifest.Artifacts), public)
}
