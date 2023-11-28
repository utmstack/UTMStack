package logservice

import (
	"sync"
)

type HttpLogService struct {
	Mutex *sync.Mutex
}

func NewHttpLogService() *HttpLogService {
	return &HttpLogService{
		Mutex: &sync.Mutex{},
	}
}
