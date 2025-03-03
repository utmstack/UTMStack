package collectors

type MacOS struct{}

func (c MacOS) Install() error {
	return nil
}

func (c MacOS) SendSystemLogs() {}

func (c MacOS) Uninstall() error {
	return nil
}
