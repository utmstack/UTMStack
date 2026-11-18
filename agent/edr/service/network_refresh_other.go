//go:build !linux

package service

import (
	"context"

	"github.com/utmstack/UTMStack/agent/edr/netblock"
)

func (*program) startNetworkSafetyRefresh(context.Context, *netblock.Manager) {}
