package domain

import "errors"

var (
	ErrUnknownDataset   = errors.New("unknown dataset")
	ErrKeepRequired     = errors.New("retention must keep records for at least a day")
	ErrColdBeforeDelete = errors.New("records must move to cold storage before they are deleted, not after")
	ErrColdNegative     = errors.New("the move to cold storage cannot be in the past")
	ErrTieringRequired  = errors.New("this instance has no cold storage: configure object storage before moving records to it")
	ErrTieringPermanent = errors.New("this dataset already moves records to cold storage and cannot go back to local-only without rebuilding the table")
	ErrColdNotReady     = errors.New("the event store has not picked up the cold storage yet")

	ErrEndpointRequired  = errors.New("the bucket URL is required")
	ErrEndpointNotURL    = errors.New("the bucket URL must be an absolute http(s) URL")
	ErrEndpointNoBucket  = errors.New("the bucket URL must include the bucket, e.g. https://s3.example.com/utmstack-cold/")
	ErrCredentialsNeeded = errors.New("both an access key and a secret key are required")
	ErrCacheNegative     = errors.New("the local cache cannot be negative")
	ErrEndpointLocked    = errors.New("records already live in the configured bucket: pointing the store at a different one would make them unreadable")
	ErrColdRefused       = errors.New("the event store could not use this bucket")
)
