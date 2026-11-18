# Quarantine platform references

The Windows purge fallback follows the handle-only Deleteat policy in Go
1.25.7, copyright 2024 The Go Authors, used under the BSD license retained in
`LICENSE.GO-BSD`. It reopens the same file object when attribute-write access
is needed, never a replacement at its former pathname.

Source: https://github.com/golang/go/blob/go1.25.7/src/internal/syscall/windows/at_windows.go

Windows delete-disposition behavior: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntddk/ns-ntddk-_file_disposition_information_ex

The owned journal/receipt protocol and its Unix metadata operations are Go code.
Windows directory durability and mapped/delete-pending behavior still need
native runtime and power-loss qualification; cross-compilation is not a pass.
