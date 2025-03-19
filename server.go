package thor

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/go-thor/thor/errors"
	"github.com/go-thor/thor/jsoncodec"
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
		return errors.ErrServerClosed
	}

	// Check if service is valid
	serviceValue := reflect.ValueOf(svc)
	if serviceValue.Kind() != reflect.Ptr || serviceValue.IsNil() {
		return errors.New(errors.ErrorCodeInvalidArgument, "service must be a non-nil pointer")
	}

	// Get service type
	serviceType := reflect.TypeOf(svc)
	serviceName := name
	if serviceName == "" {
		serviceName = reflect.Indirect(serviceValue).Type().Name()
	}
	if serviceName == "" {
		return errors.New(errors.ErrorCodeInvalidArgument, "service name cannot be empty")
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
		return errors.ErrServerClosed
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
		log.Printf("响应中的Reply数据长度: %d", len(resp.Reply))
		log.Printf("响应中的Payload数据长度: %d", len(resp.Payload))

		// 确保Response中的Payload和Reply字段已正确设置
		if len(resp.Reply) == 0 && len(resp.Payload) > 0 {
			resp.Reply = resp.Payload
		} else if len(resp.Payload) == 0 && len(resp.Reply) > 0 {
			resp.Payload = resp.Reply
		}

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
		return errors.ErrServerClosed
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
	// Apply middlewares
	var handler = func(ctx context.Context, req *Request) (*Response, error) {
		serviceName, methodName, err := parseServiceMethod(req.ServiceMethod)
		if err != nil {
			return nil, err
		}

		// Get service
		svcInterface, ok := s.serviceMap.Load(serviceName)
		if !ok {
			return nil, errors.ErrServiceNotFound
		}
		svc := svcInterface.(*service)

		// Get method
		methodType, ok := svc.method[methodName]
		if !ok {
			return nil, errors.ErrMethodNotFound
		}

		// Create argument and reply
		argv := reflect.New(methodType.ArgType.Elem()).Interface()
		replyv := reflect.New(methodType.ReplyType.Elem()).Interface()

		// 如果Args为空但Payload不为空，使用Payload作为Args
		if len(req.Args) == 0 && len(req.Payload) > 0 {
			req.Args = req.Payload
		}

		// Unmarshal argument
		err = s.codec.Unmarshal(req.Args, argv)
		if err != nil {
			return nil, fmt.Errorf("unmarshal args: %w", err)
		}

		// Call method
		err = s.call(ctx, svc, methodType, argv, replyv)
		if err != nil {
			return nil, err
		}

		// Marshal reply
		log.Printf("服务端得到的响应对象: %+v", replyv)
		replyData, err := s.codec.Marshal(replyv)
		if err != nil {
			return nil, fmt.Errorf("marshal reply: %w", err)
		}
		log.Printf("服务端序列化的响应数据长度: %d", len(replyData))
		log.Printf("服务端序列化的响应数据内容: %v", replyData)

		resp := &Response{
			ServiceMethod: req.ServiceMethod,
			Seq:           req.Seq,
			Reply:         replyData,
			Payload:       replyData,
		}
		log.Printf("服务端构造的响应: %+v", resp)
		return resp, nil
	}

	// 暂时不使用中间件，直接返回处理结果
	// 实际项目中需要根据具体的Middleware定义来实现这部分逻辑
	return handler(ctx, req)
}

// call calls the service method
func (s *DefaultServer) call(ctx context.Context, svc *service, methodType *methodType, argv, replyv interface{}) error {
	function := methodType.method.Func
	// Invoke the method, providing a new value for the reply
	returnValues := function.Call([]reflect.Value{
		svc.rcvr,
		reflect.ValueOf(ctx),
		reflect.ValueOf(argv),
	})

	// The return value for the method is an error
	errInter := returnValues[1].Interface()
	if errInter != nil {
		return errInter.(error)
	}

	return nil
}

// parseServiceMethod parses a service method string in the format "Service.Method"
func parseServiceMethod(serviceMethod string) (string, string, error) {
	dot := 0
	for i := 0; i < len(serviceMethod); i++ {
		if serviceMethod[i] == '.' {
			dot = i
			break
		}
	}
	if dot == 0 {
		return "", "", errors.New(errors.ErrorCodeInvalidArgument, "rpc: service/method request ill-formed: "+serviceMethod)
	}
	serviceName := serviceMethod[:dot]
	methodName := serviceMethod[dot+1:]
	return serviceName, methodName, nil
}

// isExported reports whether name is an exported name
func isExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}
