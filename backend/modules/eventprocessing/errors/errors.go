package errors

import "errors"

var (
	// Filter group errors — map to 400 (JHipster style).
	ErrFilterGroupIDExists  = errors.New("a new filter group cannot already have an ID")
	ErrFilterGroupIDNull    = errors.New("filter group ID must be set for update")
	ErrFilterGroupNotFound  = errors.New("filter group not found")

	// Filter errors — map to 500 (faithful Java port).
	ErrFilterIDExists   = errors.New("a new filter cannot already have an ID")
	ErrFilterIDNull     = errors.New("filter ID must be set for update")
	ErrFilterNotFound   = errors.New("filter not found")
	ErrFilterSystemOwner = errors.New("cannot update system-owned filter")

	// Pipeline errors.
	ErrPipelineNotFound    = errors.New("pipeline not found")
	ErrPipelineSystemOwner = errors.New("system pipeline cannot be deleted")
)
