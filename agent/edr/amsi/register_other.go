//go:build !windows

package amsi

// CLSID must stay byte-identical to the CLSID in the native provider DLL
// (native/utmstack_amsi.cpp) or Windows will not load the provider.
const CLSID = "{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}"

func Register(dllPath string) error { return nil }
func Unregister() error             { return nil }
