//go:build !linux

package ransomware

import "errors"

func (*canaryDir) mountKey() (string, error) {
	return "", errors.New("canary mount scoping is available only on Linux")
}
