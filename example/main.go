package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-thor/thor"
	"github.com/go-thor/thor/codec"
	"github.com/go-thor/thor/middleware"
	"github.com/go-thor/thor/transport"
)

type ExampleService struct{}

func (s *ExampleService) SayHello(ctx context.Context, req interface{}) (interface{}, error) {
	select {
	case <-time.After(1 * time.Second): // 模拟处理
		return fmt.Sprintf("Hello, %v!", req), nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled or timed out: %v", ctx.Err())
	}
}

func (s *ExampleService) Add(ctx context.Context, req interface{}) (interface{}, error) {
	nums, ok := req.([]interface{})
	if !ok || len(nums) != 2 {
		return nil, fmt.Errorf("invalid params: expected []int with 2 elements, %#v", req)
	}
	return nums[0].(float64) + nums[1].(float64), nil
}

func main() {
	jsonCodec := codec.NewJSONCodec()

	tcpTest(jsonCodec)
	httpTest(jsonCodec)
	udpTest(jsonCodec)
	websocketTest(jsonCodec)

	time.Sleep(time.Second * 10)
}

func tcpTest(code thor.Codec) {
	// TCP 服务端
	tcpTransport := transport.NewTCPTransport(code)
	serverTCP := thor.NewServer(tcpTransport, code)
	serverTCP.RegisterService(&ExampleService{})
	// serverTCP.Use(middleware.Logger())
	go serverTCP.Serve(":8080")

	go func() {
		fn := func(i int) {
			time.Sleep(time.Second * 1)
			// TCP 客户端
			clientTCP, err := thor.NewClient(tcpTransport, ":8080", code)
			if err != nil {
				fmt.Println("Tcp.NewClient.Error", i, err)
				return
			}
			// clientTCP.Use(middleware.ClientLoggerMiddleware())

			defer clientTCP.Close()

			// 设置超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			// 调用 SayHello（带超时）
			var helloResp interface{}
			err = clientTCP.Call("ExampleService.SayHello", ctx, "World via TCP "+strconv.Itoa(i), &helloResp)
			if err != nil {
				fmt.Println("Tcp.SayHello.Error：", i, err)
			} else {
				fmt.Println("Tcp.SayHello.Success：", i, helloResp)
			}

			// 调用 Add
			var addResp interface{}
			if err := clientTCP.Call("ExampleService.Add", context.Background(), []int{3, 5}, &addResp); err != nil {
				fmt.Println("Tcp.Add.Error：", i, err)
			} else {
				fmt.Println("Tcp.Add.Success：", i, addResp)
			}
		}

		for i := 0; i < 10; i++ {
			i := i

			go fn(i)
		}

		fmt.Println("")
		fmt.Println("")
	}()

}

func httpTest(code thor.Codec) {
	// HTTP 服务端
	httpTransport := transport.NewHTTPTransport(code)
	serverHTTP := thor.NewServer(httpTransport, code)
	serverHTTP.RegisterService(&ExampleService{})
	serverHTTP.Use(middleware.Logger())
	go serverHTTP.Serve(":8081")

	go func() {
		time.Sleep(time.Second * 3)

		// HTTP 客户端
		clientHTTP, err := thor.NewClient(httpTransport, ":8081", code)
		if err != nil {
			fmt.Println("Http.NewClient.Error：", err)
			return
		}
		clientHTTP.Use(middleware.ClientLoggerMiddleware())

		defer clientHTTP.Close()

		// 调用 SayHello（无超时）
		var helloRespHTTP interface{}
		if err := clientHTTP.Call("ExampleService.SayHello", context.Background(), "World via HTTP", &helloRespHTTP); err != nil {
			fmt.Println("Http.SayHello.Error：", err)
			return
		}
		fmt.Println("Http.SayHello.Success：", helloRespHTTP)

		// 调用 Add
		var addResp interface{}
		if err := clientHTTP.Call("ExampleService.Add", context.Background(), []int{1, 2}, &addResp); err != nil {
			fmt.Println("Http.Add.Error：", err)
		} else {
			fmt.Println("Http.Add.Success：", addResp)
		}

		fmt.Println("")
		fmt.Println("")
	}()
}

func udpTest(code thor.Codec) {
	// UDP 服务端
	udpTransport := transport.NewUDPTransport(code)
	serverUDP := thor.NewServer(udpTransport, code)
	serverUDP.RegisterService(&ExampleService{})
	serverUDP.Use(middleware.Logger())
	go serverUDP.Serve(":8082")

	go func() {
		time.Sleep(time.Second * 5)

		// UDP 客户端
		clientUDP, err := thor.NewClient(udpTransport, ":8082", code)
		if err != nil {
			panic(err)
		}
		clientUDP.Use(middleware.ClientLoggerMiddleware())
		defer clientUDP.Close()

		// UDP: 调用 SayHello（带超时）
		var helloRespUDP interface{}
		err = clientUDP.Call("ExampleService.SayHello", context.Background(), "World via UDP", &helloRespUDP)
		if err != nil {
			fmt.Printf("UDP.SayHello.Error: %v\n", err)
		} else {
			fmt.Println("UDP.SayHello.Success：", helloRespUDP)
		}

		// UDP: 调用 Add
		var addRespUDP interface{}
		if err := clientUDP.Call("ExampleService.Add", context.Background(), []int{5, 5}, &addRespUDP); err != nil {
			fmt.Printf("UDP.Add.Error: %v\n", err)
		} else {
			fmt.Println("UDP.Add.Success：", addRespUDP)
		}

		fmt.Println("")
		fmt.Println("")
	}()
}

func websocketTest(code thor.Codec) {
	// WebSocket 服务端
	wsTransport := transport.NewWebSocketTransport(code)
	serverWS := thor.NewServer(wsTransport, code)
	serverWS.RegisterService(&ExampleService{})
	serverWS.Use(middleware.Logger())
	go serverWS.Serve(":8083")

	go func() {
		time.Sleep(time.Second * 6)

		// WebSocket 客户端
		clientWS, err := thor.NewClient(wsTransport, ":8083", code)
		if err != nil {
			panic(err)
		}
		clientWS.Use(middleware.ClientLoggerMiddleware())
		defer clientWS.Close()

		// WebSocket: 调用 SayHello（带超时）
		var helloRespWS interface{}
		err = clientWS.Call("ExampleService.SayHello", context.Background(), "World via WebSocket", &helloRespWS)
		if err != nil {
			fmt.Printf("WebSocket.SayHello.Error: %v\n", err)
		} else {
			fmt.Println(helloRespWS)
		}

		// WebSocket: 调用 Add（复用连接）
		var addRespWS interface{}
		if err := clientWS.Call("ExampleService.Add", context.Background(), []int{3, 5}, &addRespWS); err != nil {
			fmt.Printf("WebSocket.Add.Error: %v\n", err)
		} else {
			fmt.Println("WebSocket.Add.Success：", addRespWS)
		}

		fmt.Println("")
		fmt.Println("")
	}()
}
