package rpc

import (
	"time"
	"testing"
	"reflect"
)

func TestRpc(t *testing.T) {
	var c TestClient
	var s TestServer

	server := NewServer()
	err := server.Reg(&s, &c)
	if err != nil {
		t.Fatal(err)
	}
	go server.Run("/", 11182)
	time.Sleep(time.Millisecond)

	client := NewClient("http://127.0.0.1:11182")
	err = client.Reg(&c)
	if err != nil {
		t.Fatal(err)
	}

	sum, err := c.Add(100, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	if sum != 111 {
		t.Fatal("error")
	}
}

func TestServerInvoke(t *testing.T) {
	var c TestClient
	var s TestServer

	fns := NewFnTraits(&c)

	fvs, err := NewFnValues(&s, fns)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := fvs["Larger"].Invoke(fns["Larger"].proto, []byte(`{"a":1, "b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 1 || !ret[0].(bool) {
		t.Fatal("error")
	}

	ret, err = fvs["Larger"].Invoke(fns["Larger"].proto, []byte(`{"a":2, "b":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 1 || ret[0].(bool) {
		t.Fatal("error")
	}

	ret, err = fvs["Foo"].Invoke(fns["Foo"].proto, []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ret) != 0 {
		t.Fatal("error")
	}
}

func TestFnProto(t *testing.T) {
	check := func(fns FnTraits) {
		if len(fns) != 3 {
			t.Fatal("should be 3 funcs")
		}
		add := FnProto {
			"Add",
			[]string{"a", "b", "c"},
			[]string{"int", "int", "int"},
			[]string{"sum", "err"},
			[]string{"int", "error"},
		}
		if !reflect.DeepEqual(fns["Add"].Proto(), add) {
			t.Fatal("unmatched")
		}
		larger := FnProto {
			"Larger",
			[]string{"a", "b"},
			[]string{"int", "int"},
			[]string{"larger", "err"},
			[]string{"bool", "error"},
		}
		if !reflect.DeepEqual(fns["Larger"].Proto(), larger) {
			t.Fatal("unmatched")
		}
	}

	check(NewFnTraits(&TestClient{}))
}

type TestClient struct {
	Foo    func() (error)                 `return:"err"`
	Add    func(a, b, c int)(int, error)  `args:"a,b,c" return:"sum,err"`
	Larger func(a, b int)(bool, error)    `args:"a,b" return:"larger,err"`
}

func (p *TestServer) Foo() (error) {
	return nil
}

func (p *TestServer) Add(a, b, c int) (int, error) {
	return a + b + c, nil
}

func (p *TestServer) Larger(a, b int) (bool, error) {
	return a < b, nil
}

type TestServer string
