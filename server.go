package thor

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Server struct {
	services    map[string]reflect.Value
	methods     map[string]reflect.Method
	transport   Transport
	codec       Codec
	middlewares []Middleware
	mu          sync.RWMutex
}

func NewServer(transport Transport, codec Codec) *Server {
	return &Server{
		services:  make(map[string]reflect.Value),
		methods:   make(map[string]reflect.Method),
		transport: transport,
		codec:     codec,
	}
}

func (s *Server) RegisterService(service interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sv := reflect.ValueOf(service)
	st := sv.Elem().Type()
	serviceName := st.Name()
	s.services[serviceName] = sv

	for i := 0; i < sv.Type().NumMethod(); i++ {
		method := sv.Type().Method(i)
		mt := method.Type

		if mt.NumIn() != 3 || mt.NumOut() != 2 ||
			mt.In(1) != reflect.TypeOf((*context.Context)(nil)).Elem() ||
			mt.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			panic(fmt.Sprintf("method %s.%s must have signature func(ctx context.Context, req interface{}) (interface{}, error)", serviceName, method.Name))
		}

		route := fmt.Sprintf("%s.%s", serviceName, method.Name)
		s.methods[route] = method
	}
}

func (s *Server) Use(mw Middleware) {
	s.middlewares = append(s.middlewares, mw)
}

func (s *Server) Serve(addr string) error {
	handler := s.handleRequest
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	return s.transport.ListenAndServe(addr, handler)
}

func (s *Server) handleRequest(ctx *RPCContext) error {
	req, ok := ctx.Request.(*Request)
	if !ok {
		return fmt.Errorf("invalid request format: expected rpcframework.Request, got %T: %#v", ctx.Request, ctx.Request)
	}

	route := fmt.Sprintf("%s.%s", req.ServiceName, req.MethodName)
	s.mu.RLock()
	method, exists := s.methods[route]
	service, svcExists := s.services[req.ServiceName]
	s.mu.RUnlock()

	if !exists || !svcExists {
		return fmt.Errorf("route %s not found", route)
	}

	args := []reflect.Value{service, reflect.ValueOf(ctx.Ctx), reflect.ValueOf(req.Params)}
	results := method.Func.Call(args)
	ctx.Response = results[0].Interface()
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}
	return nil
}
