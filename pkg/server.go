package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	errors2 "github.com/go-thor/thor/pkg/errors"
	"github.com/go-thor/thor/pkg/jsoncodec"
)

// service is a registered service
type service struct {
	name   string                 // name of service
	typ    reflect.Type           // receiver type
	rcvr   reflect.Value          // receiver value
	method map[string]*methodType // registered methods
}

// methodType is a registered method
type methodType struct {
	method    reflect.Method // method stub
	ArgType   reflect.Type   // argument type
	ReplyType reflect.Type   // reply type
}

// DefaultServer is the default implementation of Server
type DefaultServer struct {
	codec       Codec
	transport   Transport
	serviceMap  sync.Map // map[string]*service
	middlewares []Middleware
	mu          sync.Mutex
	closed      bool
}

// NewServer creates a new server
func NewServer(codec Codec, transport Transport) *DefaultServer {
	return &DefaultServer{
		codec:     codec,
		transport: transport,
	}
}

// Register registers a service
func (s *DefaultServer) Register(svc interface{}) error {
	return s.register(svc, "")
}

// RegisterName registers a service with a specified name
func (s *DefaultServer) RegisterName(name string, svc interface{}) error {
	return s.register(svc, name)
}

// register registers a service with the server
func (s *DefaultServer) register(svc interface{}, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors2.ErrServerClosed
	}

	// Check if service is valid
	serviceValue := reflect.ValueOf(svc)
	if serviceValue.Kind() != reflect.Ptr || serviceValue.IsNil() {
		return errors.New("service must be a non-nil pointer")
	}

	// Get service type
	serviceType := reflect.TypeOf(svc)
	serviceName := name
	if serviceName == "" {
		serviceName = reflect.Indirect(serviceValue).Type().Name()
	}
	if serviceName == "" {
		return errors.New("service name cannot be empty")
	}

	// Check if service name is valid
	if !isExported(serviceName) {
		return fmt.Errorf("service %s is not exported", serviceName)
	}

	// Create service
	svcObj := &service{
		name:   serviceName,
		typ:    serviceType,
		rcvr:   serviceValue,
		method: make(map[string]*methodType),
	}

	// Register methods
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if !isExported(method.Name) {
			continue
		}

		// Method needs 3 ins: receiver, context, *args
		if method.Type.NumIn() != 3 {
			continue
		}

		// First arg must be context.Context
		ctxType := method.Type.In(1)
		if ctxType.String() != "context.Context" {
			continue
		}

		// Second arg must be a pointer and exported
		argType := method.Type.In(2)
		if argType.Kind() != reflect.Ptr || !isExported(argType.Elem().Name()) {
			continue
		}

		// Method needs 2 outs: *reply, error
		if method.Type.NumOut() != 2 {
			continue
		}

		// First out must be a pointer and exported
		replyType := method.Type.Out(0)
		if replyType.Kind() != reflect.Ptr || !isExported(replyType.Elem().Name()) {
			continue
		}

		// Second out must be error
		errType := method.Type.Out(1)
		if errType.String() != "error" {
			continue
		}

		// Register method
		methodTyp := &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}

		svcObj.method[method.Name] = methodTyp
	}

	// Check if service has any methods
	if len(svcObj.method) == 0 {
		return fmt.Errorf("service %s has no exported methods", serviceName)
	}

	// Register service
	if _, loaded := s.serviceMap.LoadOrStore(serviceName, svcObj); loaded {
		return fmt.Errorf("service %s already registered", serviceName)
	}

	return nil
}

// Serve implements the Server interface
func (s *DefaultServer) Serve() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors2.ErrServerClosed
	}
	s.mu.Unlock()

	// 创建 JSON 编解码器用于内部消息
	jsonCodec := jsoncodec.New()

	handler := func(ctx context.Context, message []byte) ([]byte, error) {
		// Unmarshal request
		var req Request
		log.Printf("收到消息，长度: %d", len(message))
		err := jsonCodec.Unmarshal(message, &req)
		if err != nil {
			log.Printf("反序列化请求失败: %v", err)
			resp := &Response{
				Error: fmt.Sprintf("unmarshal request: %v", err),
			}
			respData, err := jsonCodec.Marshal(resp)
			if err != nil {
				log.Printf("序列化错误响应失败: %v", err)
				return nil, err
			}
			return respData, nil
		}

		log.Printf("解析请求: %+v", req)

		// Add metadata to context
		if req.Metadata != nil {
			ctx = context.WithValue(ctx, "metadata", req.Metadata)
		}

		// Add service method to context
		ctx = context.WithValue(ctx, "service_method", req.ServiceMethod)

		// Handle request
		resp, err := s.handleRequest(ctx, &req)
		if err != nil {
			log.Printf("处理请求失败: %v", err)
			resp := &Response{
				ServiceMethod: req.ServiceMethod,
				Seq:           req.Seq,
				Error:         err.Error(),
			}
			respData, err := jsonCodec.Marshal(resp)
			if err != nil {
				log.Printf("序列化错误响应失败: %v", err)
				return nil, err
			}
			return respData, nil
		}

		log.Printf("处理完成，响应: %+v", resp)

		// Marshal response
		respData, err := jsonCodec.Marshal(resp)
		if err != nil {
			log.Printf("序列化响应失败: %v", err)
			return nil, err
		}

		log.Printf("序列化完成，响应数据长度: %d", len(respData))
		return respData, nil
	}

	return s.transport.Listen("", handler)
}

// Stop stops the server
func (s *DefaultServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors2.ErrServerClosed
	}

	s.closed = true
	return s.transport.Close()
}

// Use adds middleware to the server
func (s *DefaultServer) Use(middleware ...Middleware) {
	s.middlewares = append(s.middlewares, middleware...)
}

// handleRequest handles a request
func (s *DefaultServer) handleRequest(ctx context.Context, req *Request) (*Response, error) {
	// Create response
	resp := &Response{
		ServiceMethod: req.ServiceMethod,
		Seq:           req.Seq,
		Metadata:      req.Metadata,
	}

	log.Printf("处理请求: %s, 序列号: %d", req.ServiceMethod, req.Seq)

	// Parse service and method
	serviceName, methodName, err := parseServiceMethod(req.ServiceMethod)
	if err != nil {
		resp.Error = err.Error()
		log.Printf("解析服务方法失败: %v", err)
		return resp, nil
	}

	log.Printf("服务名: %s, 方法名: %s", serviceName, methodName)

	// Get service
	svcI, ok := s.serviceMap.Load(serviceName)
	if !ok {
		resp.Error = errors2.ErrServiceNotFound.Error()
		log.Printf("找不到服务: %s", serviceName)
		return resp, nil
	}
	svcObj := svcI.(*service)

	// Get method
	methodType, ok := svcObj.method[methodName]
	if !ok {
		resp.Error = errors2.ErrMethodNotFound.Error()
		log.Printf("找不到方法: %s", methodName)
		return resp, nil
	}

	// Create arguments
	argv := reflect.New(methodType.ArgType.Elem())

	// Unmarshal payload into arguments
	err = s.codec.Unmarshal(req.Payload, argv.Interface())
	if err != nil {
		resp.Error = fmt.Sprintf("unmarshal argument: %v", err)
		log.Printf("反序列化参数失败: %v", err)
		return resp, nil
	}

	log.Printf("参数: %+v", argv.Interface())

	// Create handler for middleware chain
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Call service method
		returnValues := methodType.method.Func.Call([]reflect.Value{
			svcObj.rcvr,
			reflect.ValueOf(ctx),
			reflect.ValueOf(req),
		})

		// Check for error
		errInter := returnValues[1].Interface()
		if errInter != nil {
			return nil, errInter.(error)
		}

		return returnValues[0].Interface(), nil
	}

	// Apply middlewares
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}

	// Call handler
	reply, err := handler(ctx, argv.Interface())
	if err != nil {
		resp.Error = err.Error()
		log.Printf("调用处理器失败: %v", err)
		return resp, nil
	}

	log.Printf("响应: %+v", reply)

	// Marshal reply
	resp.Payload, err = s.codec.Marshal(reply)
	if err != nil {
		resp.Error = fmt.Sprintf("marshal reply: %v", err)
		log.Printf("序列化响应失败: %v", err)
		return resp, nil
	}

	log.Printf("响应负载长度: %d", len(resp.Payload))

	return resp, nil
}

// parseServiceMethod parses a service method string
func parseServiceMethod(serviceMethod string) (string, string, error) {
	dot := 0
	for i := 0; i < len(serviceMethod); i++ {
		if serviceMethod[i] == '.' {
			dot = i
			break
		}
	}
	if dot == 0 {
		return "", "", errors.New("rpc: service/method request ill-formed: " + serviceMethod)
	}
	return serviceMethod[:dot], serviceMethod[dot+1:], nil
}

// isExported returns true if the name is exported
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}
