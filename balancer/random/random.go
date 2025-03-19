package random

import (
	"math/rand"
	"sync"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/errors"
)

// Balancer is a random load balancer
type Balancer struct {
	instances []*pkg.ServiceInstance
	mu        sync.RWMutex
	rand      *rand.Rand
}

// New creates a new random load balancer
func New() *Balancer {
	return &Balancer{
		instances: []*pkg.ServiceInstance{},
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select selects a service instance from the instances
func (b *Balancer) Select(instances []*pkg.ServiceInstance, serviceMethod string) (*pkg.ServiceInstance, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(instances) == 0 {
		return nil, errors.ErrNoAvailableService
	}

	// Select a random instance
	index := b.rand.Intn(len(instances))
	return instances[index], nil
}

// UpdateInstances updates the instances
func (b *Balancer) UpdateInstances(instances []*pkg.ServiceInstance) {
	b.mu.Lock()
	defer b.mu.Unlock()

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
