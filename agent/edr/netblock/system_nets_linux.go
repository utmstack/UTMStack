//go:build linux

package netblock

import (
	"context"
	"errors"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// RefreshSystemNets re-resolves the manager's current endpoint configuration
// without restarting/cancelling a feed transaction. Slow lookups hold no policy
// lock. If another safety update or Reload occurred meanwhile, discard the old
// result even when the configuration's values were identical.
// A successful changed result is reconciled from the published indicator store;
// unavailable management resolution uses the existing visible fail-open policy.
func (m *Manager) RefreshSystemNets(ctx context.Context, resolve func(context.Context, config.EDRConfig) SystemNets) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if resolve == nil {
		return false, errors.New("network safety lookup is unavailable")
	}
	m.mu.Lock()
	cfg := m.cfg
	epoch := m.safetyEpoch
	needsInitial := !m.initialReconciled
	m.mu.Unlock()
	sys := resolve(ctx, cfg)
	var cached []Indicator
	var cacheErr error
	if needsInitial {
		if m.cache == nil {
			cacheErr = errors.New("network indicator cache is unavailable")
		} else if records, err := m.cache.ListIndicators(); err != nil {
			cacheErr = err
		} else {
			for _, record := range records {
				if indicator, err := ParseIndicator(record.Value, record.Type, record.Level); err == nil {
					cached = append(cached, indicator)
				}
			}
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if m.safetyEpoch != epoch {
		return false, nil
	}
	if systemNetsIdentity(sys) == m.systemNetsKey && m.enforcementHealthLocked() == nil {
		return false, nil
	}
	m.allow = BuildAllowlist(cfg, sys)
	m.configureSafetyLocked(sys)
	current, published := m.store.snapshotPublished()
	if !m.initialReconciled && !published {
		if cacheErr != nil {
			m.enforcementErr = errors.Join(cacheErr, m.enf.Clear())
			return true, m.enforcementHealthLocked()
		}
		m.store.Swap(cached)
		current, _ = m.store.snapshotPublished()
	}
	m.applyLocked(snapshotIndicators(current))
	m.initialReconciled = true
	return true, m.enforcementHealthLocked()
}

func (m *Manager) initialEnforcementHealthLocked() error {
	if m.cfg.LinuxProtection.NetworkEnforcement && !m.initialReconciled {
		return errors.New("initial network indicator policy has not reconciled")
	}
	return nil
}
