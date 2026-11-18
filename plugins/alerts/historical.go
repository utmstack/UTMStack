package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/composite"
	"github.com/threatwinds/go-sdk/store/factory"
)

type alertStores struct {
	write  store.Store
	group  store.Store
	dedup  store.Reader
	opened []*factory.Built
}

func (s *alertStores) Close() error {
	var errs []error
	for i := len(s.opened) - 1; i >= 0; i-- {
		if err := s.opened[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func openAlertStores(cfg factory.Config, opts factory.Options, open func(factory.Config, factory.Options) (*factory.Built, error)) (*alertStores, error) {
	if cfg.Primary.Engine != factory.EngineClickHouse {
		return nil, fmt.Errorf("alert groups require a ClickHouse primary store")
	}
	if cfg.Secondary != nil && !cfg.Secondary.Disabled && cfg.Secondary.Expires != "" {
		expires, err := time.Parse(time.RFC3339, cfg.Secondary.Expires)
		if err != nil {
			return nil, fmt.Errorf("invalid historical expiry: %w", err)
		}
		now := time.Now()
		if opts.Now != nil {
			now = opts.Now()
		}
		if !now.Before(expires) {
			secondary := *cfg.Secondary
			secondary.Disabled = true
			cfg.Secondary = &secondary
		}
	}
	stores := &alertStores{}
	build := func(c factory.Config) (*factory.Built, error) {
		b, err := open(c, opts)
		if err == nil {
			stores.opened = append(stores.opened, b)
		}
		return b, err
	}
	fail := func(err error) (*alertStores, error) {
		return nil, errors.Join(err, stores.Close())
	}
	built, err := build(cfg)
	if err != nil {
		return fail(err)
	}
	stores.write, stores.group, stores.dedup = built.Store, built.Store, built.Store
	if built.Composite == nil {
		return stores, nil
	}
	primaryCfg := cfg
	primaryCfg.Secondary = nil
	primary, err := build(primaryCfg)
	if err != nil {
		return fail(err)
	}
	stores.group = primary.Store
	if !secondaryServesAlerts(cfg) {
		return stores, nil
	}
	secondary, err := build(factory.Config{Primary: cfg.Secondary.Engine})
	if err != nil {
		return fail(err)
	}
	stores.dedup, err = newHistoricalCounter(cfg, primary.Store, secondary.Store, opts)
	if err != nil {
		return fail(err)
	}
	return stores, nil
}

func secondaryServesAlerts(cfg factory.Config) bool {
	if cfg.Secondary == nil || cfg.Secondary.Disabled {
		return false
	}
	for _, dataset := range cfg.Secondary.Datasets {
		if dataset == string(datasetAlerts) {
			return true
		}
	}
	return false
}

func newHistoricalCounter(cfg factory.Config, primary, secondary store.Store, opts factory.Options) (store.Store, error) {
	if cfg.Secondary == nil {
		return nil, fmt.Errorf("historical alert deduplication requires a secondary store")
	}
	cutover, err := time.Parse(time.RFC3339, cfg.Secondary.Cutover)
	if err != nil {
		return nil, fmt.Errorf("invalid historical cutover: %w", err)
	}
	var expires time.Time
	if cfg.Secondary.Expires != "" {
		expires, err = time.Parse(time.RFC3339, cfg.Secondary.Expires)
		if err != nil {
			return nil, fmt.Errorf("invalid historical expiry: %w", err)
		}
	}
	datasets := make([]store.Dataset, len(cfg.Secondary.Datasets))
	for i, dataset := range cfg.Secondary.Datasets {
		datasets[i] = store.Dataset(dataset)
	}
	cc := composite.Config{
		Primary: primary, Secondary: secondary, Cutover: cutover, SecondaryExpires: expires,
		SecondaryTenant: cfg.Secondary.Tenant, SecondaryDatasets: datasets,
		TolerateSecondary: cfg.Secondary.TolerateFailure, OnSecondaryError: opts.OnSecondaryError, Now: opts.Now,
	}
	if prefix := cfg.Secondary.FieldPrefix; prefix != nil {
		if prefix.Primary != "log" || prefix.Secondary == "" || prefix.Secondary == prefix.Primary {
			return nil, fmt.Errorf("historical alert fieldPrefix requires primary log and a different secondary field")
		}
		toSecondary, fromSecondary := composite.PrefixMap(prefix.Primary, prefix.Secondary)
		cc.SecondaryField = func(field string) string {
			if field == "lastEvent."+prefix.Primary || strings.HasPrefix(field, "lastEvent."+prefix.Primary+".") {
				return "lastEvent." + toSecondary(strings.TrimPrefix(field, "lastEvent."))
			}
			return toSecondary(field)
		}
		cc.SecondaryFieldBack = fromSecondary
		cc.SecondaryDocument = composite.RenameDocKey(prefix.Secondary, prefix.Primary)
	}
	return composite.New(cc)
}
