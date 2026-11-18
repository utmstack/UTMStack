//go:build !linux

package main

func configureGoMemoryLimit() int64 { return 0 }
