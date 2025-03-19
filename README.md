# Thor RPC Framework

Thor is a high-performance RPC (Remote Procedure Call) framework for Go applications, focused on simplicity, performance, and extensibility. It's designed to handle high concurrency workloads and provides a robust foundation for building distributed systems.

## Features

- **High Performance**: Optimized for high throughput and low latency
- **Protocol Buffers**: First-class support for Protocol Buffers as IDL
- **Multiple Codecs**: Support for Protocol Buffers, JSON, and extensible for other formats
- **Multiple Transports**: TCP, HTTP, and UDP transport layers
- **Middleware Support**: Pluggable middleware architecture for cross-cutting concerns
- **Load Balancing**: Built-in support for various load balancing strategies
- **Service Discovery**: Flexible service discovery mechanism
- **Timeout Control**: Fine-grained timeout controls
- **Metadata**: Support for passing metadata between client and server
- **Error Handling**: Comprehensive error handling system

## Installation

```bash
go get github.com/go-thor/thor
```

## Quick Start

### Define your service using Protocol Buffers

```protobuf
syntax = "proto3";

package example;
option go_package = "github.com/example/service";

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloResponse);
}

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}
```

### Generate code

```bash
protoc --proto_path=. --go_out=. --thor_out=. example.proto
```

### Server Implementation

```go
package main

import (
	"context"
	"log"
	"net"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/codec/protobuf"
	"github.com/go-thor/thor/transport/tcp"
	"github.com/example/service"
)

type GreeterService struct{}

func (s *GreeterService) SayHello(ctx context.Context, req *service.HelloRequest) (*service.HelloResponse, error) {
	return &service.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func main() {
	// Create a TCP transport
	transport := tcp.New()

	// Create a Protobuf codec
	codec := protobuf.New()

	// Create a server
	server := pkg.NewServer(codec, transport)

	// Register the service
	service.RegisterGreeterServer(server, &GreeterService{})

	// Start the server
	if err := server.Serve(); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
```

### Client Implementation

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-thor/thor/pkg"
	"github.com/go-thor/thor/codec/protobuf"
	"github.com/go-thor/thor/transport/tcp"
	"github.com/example/service"
)

func main() {
	// Create a TCP transport
	transport := tcp.New()

	// Create a Protobuf codec
	codec := protobuf.New()

	// Create a client
	client := pkg.NewClient(codec, transport)
	defer client.Close()

	// Create a service client
	greeter := service.NewGreeterClient(client)

	// Call the service
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := greeter.SayHello(ctx, &service.HelloRequest{Name: "Thor"})
	if err != nil {
		log.Fatalf("Failed to call service: %v", err)
	}

	fmt.Println(response.Message)
}
```

## Middleware

Thor supports middleware for cross-cutting concerns such as logging, metrics, tracing, and authentication.

### Logging Middleware

```go
import "github.com/go-thor/thor/middleware/logging"

// Create a server with logging middleware
server := pkg.NewServer(codec, transport)
server.Use(logging.New())
```

### Metrics Middleware

```go
import "github.com/go-thor/thor/middleware/metrics"

// Create a server with metrics middleware
server := pkg.NewServer(codec, transport)
server.Use(metrics.New())
```

### Tracing Middleware

```go
import "github.com/go-thor/thor/middleware/tracing"

// Create a server with tracing middleware
server := pkg.NewServer(codec, transport)
server.Use(tracing.New())
```

### JWT Authentication Middleware

```go
import "github.com/go-thor/thor/middleware/auth"

// Create a server with JWT authentication middleware
server := pkg.NewServer(codec, transport)
server.Use(auth.NewJWT(auth.WithSecret([]byte("your-secret"))))
```

## Transport Layers

Thor supports multiple transport layers for communication between clients and servers.

### TCP Transport

```go
import "github.com/go-thor/thor/transport/tcp"

// Create a TCP transport
transport := tcp.New()
```

### HTTP Transport

```go
import "github.com/go-thor/thor/transport/http"

// Create an HTTP transport
transport := http.New()
```

### UDP Transport

```go
import "github.com/go-thor/thor/transport/udp"

// Create a UDP transport
transport := udp.New()
```

## Codecs

Thor supports multiple codecs for serialization and deserialization of messages.

### Protocol Buffers Codec

```go
import "github.com/go-thor/thor/codec/protobuf"

// Create a Protocol Buffers codec
codec := protobuf.New()
```

### JSON Codec

```go
import "github.com/go-thor/thor/codec/json"

// Create a JSON codec
codec := json.New()
```

## Service Discovery

Thor provides service discovery mechanisms to allow clients to find service instances.

```go
import "github.com/go-thor/thor/discovery/memory"

// Create an in-memory service discovery
discovery := memory.New()

// Register a service
discovery.Register("greeter", "localhost:50051", nil)

// Get service instances
instances, err := discovery.GetService("greeter")
```

## Load Balancing

Thor includes load balancing strategies to distribute requests across multiple service instances.

### Random Load Balancer

```go
import "github.com/go-thor/thor/balancer/random"

// Create a random load balancer
balancer := random.New()
```

### Round Robin Load Balancer

```go
import "github.com/go-thor/thor/balancer/round_robin"

// Create a round robin load balancer
balancer := round_robin.New()
```

## Error Handling

Thor provides a robust error handling system for dealing with various types of errors.

```go
import "github.com/go-thor/thor/pkg/errors"

// Check error type
if errors.Is(err, errors.ErrTimeout) {
    // Handle timeout error
} else if errors.Is(err, errors.ErrServerClosed) {
    // Handle server closed error
}

// Create custom error
err := errors.New(errors.ErrorCodeInvalidArgument, "invalid argument")

// Wrap error
err = errors.Wrap(errors.ErrorCodeUnknown, originalErr, "context message")
```

## License

Thor is licensed under the MIT License. See [LICENSE](LICENSE) for the full license text.