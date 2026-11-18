//go:build linux

package ransomware

import (
	"context"
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
)

type disabledLinuxFeed struct{}

// Linux activation requires explicit mount/roots via NewFanotifyFeed.
func NewFeed() FileActivityFeed                                      { return disabledLinuxFeed{} }
func (disabledLinuxFeed) Run(context.Context, func(FileEvent)) error { return fanotify.ErrDisabled }
func (disabledLinuxFeed) Health() error                              { return fanotify.ErrDisabled }
