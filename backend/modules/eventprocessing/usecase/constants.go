package usecase

import "time"

const (
	MaxPlaygroundBodyBytes  int64         = 1 << 20 // 1 MiB
	SemaphoreAcquireTimeout time.Duration = 30 * time.Second
	PlaygroundUserFilename  string        = "playground-user.yaml"
)
