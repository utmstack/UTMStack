//go:build !linux

package fanotify

func NewPermissionBroker(Config, Scanner, func(Outcome)) (*Monitor, error) {
	return nil, ErrUnsupported
}
