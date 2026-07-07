//go:build windows

package amsi

import "golang.org/x/sys/windows/registry"

// CLSID must stay byte-identical to the CLSID in the native provider DLL
// (native/utmstack_amsi.cpp) or Windows will not load the provider.
const CLSID = "{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}"

// Register writes the COM InprocServer32 for the provider DLL and registers it
// as an AMSI provider under HKLM.
func Register(dllPath string) error {
	inproc := `SOFTWARE\Classes\CLSID\` + CLSID + `\InprocServer32`
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, inproc, registry.SET_VALUE)
	if err != nil {
		return err
	}
	if err := k.SetStringValue("", dllPath); err != nil {
		k.Close()
		return err
	}
	if err := k.SetStringValue("ThreadingModel", "Both"); err != nil {
		k.Close()
		return err
	}
	k.Close()

	prov := `SOFTWARE\Microsoft\AMSI\Providers\` + CLSID
	pk, _, err := registry.CreateKey(registry.LOCAL_MACHINE, prov, registry.SET_VALUE)
	if err != nil {
		return err
	}
	pk.Close()
	return nil
}

func Unregister() error {
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\AMSI\Providers\`+CLSID)
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\CLSID\`+CLSID+`\InprocServer32`)
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\CLSID\`+CLSID)
	return nil
}
