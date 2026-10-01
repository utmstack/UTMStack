//go:build !windows && !linux

package ransomware

// t1490Rules is empty off Windows and Linux: the recovery-tampering sensor
// only exists where the feed and command rules are implemented, but the
// package-level var must exist on every platform so rules.go compiles.
var t1490Rules = []t1490Rule{}
