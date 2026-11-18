//go:build darwin || linux

package bytecode

import "golang.org/x/sys/unix"

// ClamAV copies uname's four fields into its bytecode environment record.
// Keep the terminating zero outside the copied field's 64-byte payload.
func environmentUname() (sysname, release, version, machine string) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return "", "", "", ""
	}
	return utsString(u.Sysname[:]), utsString(u.Release[:]), utsString(u.Version[:]), utsString(u.Machine[:])
}

func utsString[T ~int8 | ~uint8](field []T) string {
	b := make([]byte, 0, len(field))
	for _, c := range field {
		if c == 0 || len(b) == 64 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}
