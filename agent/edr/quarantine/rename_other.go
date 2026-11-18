//go:build !linux && !darwin && !windows

package quarantine

import "errors"

func renameNoReplace(src, dst string) error { return renameError(src, dst, errors.ErrUnsupported) }
