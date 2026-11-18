//go:build !windows

package amsi

import (
	"context"
	"errors"
)

const PipeName = `\\.\pipe\utmstack_edr_amsi`

type PipeServer struct{}

var ErrPipeNotReady = errors.New("AMSI local pipe is unavailable on this platform")

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{} }

func (p *PipeServer) Run(ctx context.Context) { <-ctx.Done() }
func (p *PipeServer) Health() error           { return ErrPipeNotReady }
