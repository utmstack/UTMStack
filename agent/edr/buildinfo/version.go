// Package buildinfo identifies the native EDR distribution independently of
// signature generations and matcher cache versions.
package buildinfo

import "runtime"

// Version changes when the agent must fetch a new EDR executable. Keep this
// separate from the platform version so an older tracked 12.0.0 EDR is updated.
const Version = "12.0.125-native"

// ImplementationVersion is shared by the scanner and the agent's updater.
// A scanner implementation change must also advance Version so installed
// agents fetch the matching executable instead of retaining an older core.
const ImplementationVersion = "native-go-v293"

type Info struct {
	PackageVersion string `json:"package_version"`
	Implementation string `json:"implementation"`
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	GoVersion      string `json:"go_version"`
}

func Current(implementation string) Info {
	return Info{
		PackageVersion: Version,
		Implementation: implementation,
		OS:             runtime.GOOS,
		Architecture:   runtime.GOARCH,
		GoVersion:      runtime.Version(),
	}
}
