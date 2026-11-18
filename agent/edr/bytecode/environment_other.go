//go:build !darwin && !linux

package bytecode

func environmentUname() (sysname, release, version, machine string) {
	return "", "", "", ""
}
