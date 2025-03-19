package memory

import (
	"sync"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/pkg/errors"
)

// Discovery is an in-memory service discovery
type Discovery struct {
	services map[string][]*pkg.ServiceInstance
	watchers map[string][]chan []*pkg.ServiceInstance
	mu       sync.RWMutex
}

// New creates a new in-memory service discovery
func New() *Discovery {
	return &Discovery{
		services: make(map[string][]*pkg.ServiceInstance),
		watchers: make(map[string][]chan []*pkg.ServiceInstance),
	}
}

// Register registers a service with the discovery
func (d *Discovery) Register(serviceName, addr string, metadata map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if the service instance already exists
	instances, ok := d.services[serviceName]
	if ok {
		for _, instance := range instances {
			if instance.Addr == addr {
				// Update metadata
				instance.Metadata = metadata
				// Notify watchers
				d.notifyWatchers(serviceName)
				return nil
			}
		}
	}

	// Add the service instance
	instance := &pkg.ServiceInstance{
		ServiceName: serviceName,
		Addr:        addr,
		Metadata:    metadata,
	}

	d.services[serviceName] = append(d.services[serviceName], instance)

	// Notify watchers
	d.notifyWatchers(serviceName)

	return nil
}

// Unregister unregisters a service from the discovery
func (d *Discovery) Unregister(serviceName, addr string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	instances, ok := d.services[serviceName]
	if !ok {
		return errors.ErrServiceNotFound
	}

	for i, instance := range instances {
		if instance.Addr == addr {
			// Remove the service instance
			d.services[serviceName] = append(instances[:i], instances[i+1:]...)

			// If there are no more instances, remove the service
			if len(d.services[serviceName]) == 0 {
				delete(d.services, serviceName)
			}

			// Notify watchers
			d.notifyWatchers(serviceName)

			return nil
		}
	}

	return errors.ErrServiceNotFound
}

// GetService gets the instances of a service
func (d *Discovery) GetService(serviceName string) ([]*pkg.ServiceInstance, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instances, ok := d.services[serviceName]
	if !ok {
		return nil, errors.ErrServiceNotFound
	}

	// Return a copy of the instances
	result := make([]*pkg.ServiceInstance, len(instances))
	for i, instance := range instances {
		result[i] = &pkg.ServiceInstance{
			ServiceName: instance.ServiceName,
			Addr:        instance.Addr,
			Metadata:    make(map[string]string),
		}
		for k, v := range instance.Metadata {
			result[i].Metadata[k] = v
		}
	}

	return result, nil
}

// Watch watches for changes of a service
func (d *Discovery) Watch(serviceName string) (chan []*pkg.ServiceInstance, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create a channel for the watcher
	ch := make(chan []*pkg.ServiceInstance, 10)

	// Add the watcher
	d.watchers[serviceName] = append(d.watchers[serviceName], ch)

	// Send the current instances
	instances, ok := d.services[serviceName]
	if ok {
		// Create a copy of the instances
		result := make([]*pkg.ServiceInstance, len(instances))
		for i, instance := range instances {
			result[i] = &pkg.ServiceInstance{
				ServiceName: instance.ServiceName,
				Addr:        instance.Addr,
				Metadata:    make(map[string]string),
			}
			for k, v := range instance.Metadata {
				result[i].Metadata[k] = v
			}
		}

		// Send the instances to the watcher
		go func() {
			ch <- result
		}()
	}

	return ch, nil
}

// notifyWatchers notifies all watchers of a service
func (d *Discovery) notifyWatchers(serviceName string) {
	watchers, ok := d.watchers[serviceName]
	if !ok {
		return
	}

	instances, ok := d.services[serviceName]
	if !ok {
		instances = []*pkg.ServiceInstance{}
	}

	// Create a copy of the instances
	result := make([]*pkg.ServiceInstance, len(instances))
	for i, instance := range instances {
		result[i] = &pkg.ServiceInstance{
			ServiceName: instance.ServiceName,
			Addr:        instance.Addr,
			Metadata:    make(map[string]string),
		}
		for k, v := range instance.Metadata {
			result[i].Metadata[k] = v
		}
	}

	// Notify all watchers
	for _, ch := range watchers {
		go func(ch chan []*pkg.ServiceInstance) {
			// Try to send the instances, but don't block
			select {
			case ch <- result:
			case <-time.After(time.Second):
				// Timed out, the watcher might be slow
			}
		}(ch)
	}
}

// Make sure Discovery implements pkg.Discovery
var _ pkg.Discovery = (*Discovery)(nil)
