# Thor RPC Framework Examples

This directory contains examples of how to use the Thor RPC framework.

## Greeter Example

A simple RPC service that demonstrates the basic functionality of Thor.

### Running the Example

1. Generate the protobuf code:

```bash
# From the project root
cd examples/greeter
protoc --proto_path=proto --go_out=paths=source_relative:proto --thor_out=paths=source_relative:proto proto/greeter.proto
```

2. Run the server:

```bash
go run server/main.go
```

3. In a separate terminal, run the client:

```bash
go run client/main.go
```

You should see the client output showing successful RPC calls to the server.

## Other Examples

More examples showcasing different features of Thor:

- **HTTP Transport**: Demonstrates using HTTP as the transport layer
- **UDP Transport**: Demonstrates using UDP as the transport layer
- **Load Balancing**: Shows how to use the load balancing feature
- **Service Discovery**: Demonstrates service discovery
- **Middleware**: Examples of using different middleware
- **Error Handling**: How to handle errors properly
- **Timeout Control**: Shows timeout control features
- **Metadata**: How to use metadata in RPC calls