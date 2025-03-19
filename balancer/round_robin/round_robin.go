package round_robin

import (
	"sync"
	"sync/atomic"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/pkg/errors"
)

// Balancer is a round-robin load balancer
type Balancer struct {
	instances []*pkg.ServiceInstance
	mu        sync.RWMutex
	next      uint64
}

// New creates a new round-robin load balancer
func New() *Balancer {
	return &Balancer{
		instances: []*pkg.ServiceInstance{},
		next:      0,
	}
}

// Select selects a service instance from the instances
func (b *Balancer) Select(instances []*pkg.ServiceInstance, serviceMethod string) (*pkg.ServiceInstance, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(instances) == 0 {
		return nil, errors.ErrNoAvailableService
	}

	// Get the next instance in a round-robin fashion
	next := atomic.AddUint64(&b.next, 1) - 1
	index := next % uint64(len(instances))
	return instances[index], nil
}

// UpdateInstances updates the instances
func (b *Balancer) UpdateInstances(instances []*pkg.ServiceInstance) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset the counter if we're updating instances
	atomic.StoreUint64(&b.next, 0)

	// Make a copy of the instances
	b.instances = make([]*pkg.ServiceInstance, len(instances))
	for i, instance := range instances {
		b.instances[i] = &pkg.ServiceInstance{
			ServiceName: instance.ServiceName,
			Addr:        instance.Addr,
			Metadata:    make(map[string]string),
		}
		for k, v := range instance.Metadata {
			b.instances[i].Metadata[k] = v
		}
	}
}

// Make sure Balancer implements pkg.Balancer
var _ pkg.Balancer = (*Balancer)(nil)
