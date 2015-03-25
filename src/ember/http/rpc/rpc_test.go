package rpc

import (
	"sync"
	"testing"
	"time"
)

var testServerOnce sync.Once
var testClientOnce sync.Once

var testServer *Server
var testClient *Client

func newTestServer() *Server {
	f := func() {
		testServer = NewServer("Router")
		go testServer.Run()
	}

	testServerOnce.Do(f)

	return testServer
}

func newTestClient() *Client {
	f := func() {
		testClient = NewClient("http://127.0.0.1:11182/")
	}

	testClientOnce.Do(f)

	return testClient
}

type Integer int;

func (p *Integer) Larger(a, b int) (bool, error) {
	return a < b, nil
}

func (p *Integer) Add(a, b int) (int, error) {
	return a + b, nil
}

func TestRpc0(t *testing.T) {
	println("-------------------- B")
	var a Integer

	s := newTestServer()
	s.RegisterObj(&a)
	time.Sleep(time.Second * 1)

	c := newTestClient()
	
	type B struct {
		Larger func(a, b int)(bool, error)	
		Add func(a, b int)(int, error)	
	}
	var b B

	if err := c.MakeRpcObj(&b); err != nil {
		t.Fatal(err)
	}
	
	if larger, err := b.Larger(10, 1); err != nil {
		t.Fatal(err)
	} else {
		println(larger)	
	}
	println("-------------------- E")


}

