//go:build linux

package service

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/netblock"
	"github.com/utmstack/UTMStack/shared/logger"
)

const linuxNetworkSafetyRefreshInterval = 15 * time.Second

func (p *program) startNetworkSafetyRefresh(ctx context.Context, manager *netblock.Manager) {
	p.goNative("network-safety", func() {
		ticker := time.NewTicker(linuxNetworkSafetyRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := manager.RefreshSystemNets(ctx, resolveSystemNetsContext); err != nil && !errors.Is(err, context.Canceled) {
					logger.Error("UTMStack EDR: network safety refresh: %v", err)
				}
			}
		}
	})
}
