package service

import "github.com/utmstack/UTMStack/agent/edr/netblock"

type networkHealth struct{ manager *netblock.Manager }

func (h networkHealth) Health() error { return h.manager.EnforcementHealth() }
