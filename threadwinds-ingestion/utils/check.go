package utils

import (
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	wait = 5 * time.Second
)

func InfiniteRetryIfXError(f func() error, exceptions ...string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exceptions...) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, nil)
				xErrorWasLogged = true
			}
			time.Sleep(wait)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
