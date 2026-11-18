//go:build !linux && !darwin

package ransomware

import "os"

const canaryPinRequired = false

func duplicateCanaryFile(*os.File) (*os.File, error) { return nil, nil }

// Windows' ETW feed remains path based; Unix file IDs are not inferred from
// Win32FileAttributeData or a pathname after the event.
func canaryFileID(os.FileInfo) (canaryFileIdentity, bool) {
	return canaryFileIdentity{}, false
}
