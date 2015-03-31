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
		testServer = NewServer()
		go testServer.Run(11182)
	}
	testServerOnce.Do(f)
	return testServer
}

func newTestClient() *Client {
	f := func() {
		testClient = NewClient("http://127.0.0.1:11182")
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

func (p *Integer) Trait() map[string][]string {
	//"Larger": []string{"a", "b"},
	//"Add": []string{"a", "b"},
	m := make(map[string][]string)
	m["Larger"] = []string{"a", "b"}
	m["Add"] = []string{"a", "b"}
	return m
}

type TestStruct struct {
	x int
	y int

}

func TestRpc(t *testing.T) {
	var a Integer

	s := newTestServer()
	s.Register(&a)
	time.Sleep(time.Millisecond)

	c := newTestClient()
	
	type B struct {
		Larger func(a, b int)(bool, error)	
		Add func(a, b int)(int, error)	
	}
	var b B

	if err := c.MakeRpc(&b, &a); err != nil {
		t.Fatal(err)
	}

	ret, err := b.Add(100, 10)
	if err != nil {
		t.Fatal(err)
	}
	if ret != 110 {
		t.Fatal(err)
	}
}
