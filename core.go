package thor

import (
	"context"
	"time"
)

// Request represents an RPC request
type Request struct {
	// ServiceMethod is the format of "service.method"
	ServiceMethod string
	// Metadata contains additional information about the request
	Metadata map[string]string
	// Payload contains the actual data of the request
	Payload []byte
	// Args contains the argument data
	Args []byte
	// Seq is the sequence number of the request
	Seq uint64
}

// Response represents an RPC response
type Response struct {
	// ServiceMethod is the format of "service.method"
	ServiceMethod string
	// Metadata contains additional information about the response
	Metadata map[string]string
	// Payload contains the actual data of the response
	Payload []byte
	// Reply contains the reply data
	Reply []byte
	// Error contains the error message if the call fails
	Error string
	// Seq is the sequence number of the request, used to match the response
	Seq uint64
}

// Client defines the interface for an RPC client
type Client interface {
	// Call invokes the named function, waits for it to complete, and returns its error status
	Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error
	// CallWithMetadata is like Call but with additional metadata
	CallWithMetadata(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, metadata map[string]string) error
	// Go invokes the function asynchronously
	Go(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call
	// Close closes the client
	Close() error
	// Use adds middleware to the client
	Use(middleware ...Middleware)
}

// Server defines the interface for an RPC server
type Server interface {
	// Register registers a service
	Register(svc interface{}) error
	// RegisterName registers a service with a specified name
	RegisterName(name string, svc interface{}) error
	// Serve starts serving requests
	Serve() error
	// Stop stops the server
	Stop() error
	// Use adds middleware to the server
	Use(middleware ...Middleware)
}

// Call represents an active RPC
type Call struct {
	// ServiceMethod is the name of the service and method to call
	ServiceMethod string
	// Args holds the arguments to the function
	Args interface{}
	// Reply holds the reply from the function
	Reply interface{}
	// Error holds the error from the call
	Error error
	// Done is a channel that is closed when the call is complete
	Done chan *Call
	// Metadata contains additional information about the call
	Metadata map[string]string
	// Context holds the context for the call
	Context context.Context
	// Seq is the sequence number of the call
	Seq uint64
}

// Codec defines how to encode and decode messages
type Codec interface {
	// Marshal converts a Go object into bytes
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal converts bytes back into a Go object
	Unmarshal(data []byte, v interface{}) error
	// Name returns the name of the codec
	Name() string
}

// Transport defines the interface for network transport
type Transport interface {
	// Send sends the message to the specified address
	Send(ctx context.Context, addr string, message []byte) ([]byte, error)
	// Listen starts listening for incoming messages
	Listen(addr string, handler func(ctx context.Context, message []byte) ([]byte, error)) error
	// Close closes the transport
	Close() error
	// Name returns the name of the transport
	Name() string
}

// ServiceDesc represents an RPC service's specification
type ServiceDesc struct {
	// ServiceName is the name of the service
	ServiceName string
	// HandlerType is the type of the handler
	HandlerType interface{}
	// Methods contains the methods of the service
	Methods []MethodDesc
}

// MethodDesc represents an RPC method's specification
type MethodDesc struct {
	// MethodName is the name of the method
	MethodName string
	// Handler is the method handler
	Handler func(srv interface{}, ctx context.Context, dec func(interface{}) error) (interface{}, error)
}

// Middleware defines a function that can be used as middleware
type Middleware func(ctx context.Context, req *Request, next func(ctx context.Context, req *Request) (*Response, error)) (*Response, error)

// Handler defines the handler function for middleware
type Handler func(ctx context.Context, req interface{}) (interface{}, error)

// Options defines options for client and server
type Options struct {
	// Timeout is the timeout for a client call
	Timeout time.Duration
	// Codec is the codec to use
	Codec Codec
	// Transport is the transport to use
	Transport Transport
	// Middlewares is the middlewares to use
	Middlewares []Middleware
	// Discovery is the service discovery to use
	Discovery Discovery
	// Balancer is the load balancer to use
	Balancer Balancer
}

// Discovery defines the interface for service discovery
type Discovery interface {
	// Register registers a service with the discovery
	Register(serviceName, addr string, metadata map[string]string) error
	// Unregister unregisters a service from the discovery
	Unregister(serviceName, addr string) error
	// GetService gets the instances of a service
	GetService(serviceName string) ([]*ServiceInstance, error)
	// Watch watches for changes of a service
	Watch(serviceName string) (chan []*ServiceInstance, error)
}

// ServiceInstance represents an instance of a service
type ServiceInstance struct {
	// ServiceName is the name of the service
	ServiceName string
	// Addr is the address of the service
	Addr string
	// Metadata contains additional information about the service
	Metadata map[string]string
}

// Balancer defines the interface for load balancing
type Balancer interface {
	// Select selects a service instance from the instances
	Select(instances []*ServiceInstance, serviceMethod string) (*ServiceInstance, error)
	// UpdateInstances updates the instances
	UpdateInstances(instances []*ServiceInstance)
}
