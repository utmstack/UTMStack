package mcp

// licenseFlags is a tiny indirection over billing.LicenseService so Gate.check
// stays decoupled from the billing package and easier to fake in tests.
type licenseFlags struct {
	isEnterprise func() bool
	isMSSP       func() bool
}

func (d *Deps) licenseFlags() licenseFlags {
	if d == nil || d.Billing == nil {
		return licenseFlags{
			isEnterprise: func() bool { return false },
			isMSSP:       func() bool { return false },
		}
	}
	lic := d.Billing.License()
	return licenseFlags{
		isEnterprise: func() bool { return lic.Current().IsEnterprise() },
		isMSSP:       func() bool { return lic.Current().IsMSSP() },
	}
}
