package xclient

import "time"

type GrpcRegistryDiscovery struct {
	*MultiServersDiscovery
	registry string
	timeout time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewGrpcRegistryDiscovery(registerAddr string, timeout time.Duration)*GrpcRegistryDiscovery  {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &GrpcRegistryDiscovery{
		MultiServersDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry: registerAddr,
		timeout: timeout,
	}
	return d
}

