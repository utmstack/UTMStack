package serv

import (
	"github.com/kardianos/service"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func getConfigServ(name, displayName, description string) *service.Config {
	svcConfig := &service.Config{
		Name:        name,
		DisplayName: displayName,
		Description: description,
	}

	return svcConfig
}
