package snowrpc

import (
	"testing"
	"time"
)

type RPC struct{}

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Reply struct {
	C int `json:"c"`
}

func (r *RPC) Add(req *Request, rep *Reply) error {
	rep.C = req.A + req.B
	return nil
}

func TestRPC(t *testing.T) {
	srv := NewServer()
	srv.RegisterName("calc", &RPC{})
	go srv.ListenAndServe("tcp", ":12345")

	time.Sleep(100 * time.Millisecond)

	cli, err := Dial("tcp", "localhost:12345")
	if err != nil {
		t.Fatal(err)
	}
	req := &Request{A: 1, B: 2}
	var rep Reply
	err = cli.Call("calc.Add", req, &rep)
	if err != nil {
		t.Fatal(err)
	}
	if rep.C != req.A+req.B {
		t.Fatalf("SnowRPC result is wrong")
	}
}
