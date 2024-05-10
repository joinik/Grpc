package registry

import (
	"sort"
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
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_Grpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *GrpcRegistry {
	return &GrpcRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultGrpcRegister = New(defaultTimeout)

// 添加服务实例
func (r *GrpcRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now() //if exists, update staret itme to keep alive
	}
}

// 返回可用的服务列表，如果存在超时的服务，则删除
func (r *GrpcRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}
