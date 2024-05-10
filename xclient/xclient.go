package xclient

import (
	"Grpc"
	"Grpc/service"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *service.Option
	mu      sync.Mutex //protect following
	clients map[string]*Grpc.Client
}

// Close implements io.Closer.
func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		// I have no idea how to deal with error, just ignore it
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil

}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *service.Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*Grpc.Client)}
}
