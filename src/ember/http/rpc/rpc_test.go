package rpc

import (
	"testing"
	"time"
)

func TestRpc(t *testing.T) {
	var a Integer

	s := NewServer()
	s.Reg(&a)

	go s.Run(11182)
	time.Sleep(time.Millisecond)

	c := NewClient("http://127.0.0.1:11182")

	type B struct {
		Larger func(a, b int)(bool, error)	
		Add func(a, b int)(int, error)	
	}
	var b B

	if err := c.Reg(&b, &a); err != nil {
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

func (p *Integer) Larger(a, b int) (bool, error) {
	return a < b, nil
}

func (p *Integer) Add(a, b int) (int, error) {
	return a + b, nil
}

func (p *Integer) Trait() map[string][]string {
	m := map[string][]string {
	"Larger": []string{"a", "b"},
	"Add": []string{"a", "b"},
	}
	return m
}

type Integer int;

