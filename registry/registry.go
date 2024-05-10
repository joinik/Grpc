package registry

import (
	"sync"
	"time"
)

// GrpcRegistry is a simple register center, provide following functions
// add a server and receive heartbeat to keep it alive.
// returns all alive servers and delete dead servers sync simultaneously.
type GrpcRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr string 
	start time.Time

}

const (
	defaultPath = "/_Grpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *GrpcRegistry  {
	return &GrpcRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultGrpcRegister = New(defaultTimeout)

