package snowrpc_test

import (
	"github.com/riobard/go-snowrpc"
	"log"
	"net"
)

func ExampleDial() {
	cli, err := snowrpc.Dial("tcp", "localhost:12345")
	if err != nil {
		log.Fatal(err)
	}
	request := "hello"
	var reply struct {
		Msg string
	}
	cli.Call("DemoService.Echo", request, &reply)
}

func ExampleNewClient() {
	conn, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		log.Fatal(err)
	}
	cli := snowrpc.NewClient(conn)
	request := "hello"
	var reply struct {
		Msg string
	}
	cli.Call("DemoService.Echo", request, &reply)
}

func CalcService() interface{} {
	return nil
}

func ExampleServer() {
	srv := snowrpc.NewServer()
	srv.RegisterName("calc", CalcService())
	srv.ListenAndServe("tcp", "locahost:12345")
}
