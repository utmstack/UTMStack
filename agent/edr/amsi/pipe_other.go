//go:build !windows

package amsi

import "context"

const PipeName = `\\.\pipe\utmstack_edr_amsi`

type PipeServer struct{}

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{} }

func (p *PipeServer) Run(ctx context.Context) { <-ctx.Done() }
