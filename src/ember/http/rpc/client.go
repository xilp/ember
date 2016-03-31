package rpc

import (
	"errors"
	"fmt"
	"strings"
	"ember/measure"
)

func (p *Client) Reg(obj interface{}) (err error) {
	return p.reg("", obj)
}

func (p *Client) List() (fns []FnProto) {
	return p.fns.List()
}

func (p *Client) Call(name string, args []string) (ret []interface{}, err error) {
	fn := p.fns[name]
	if fn == nil {
		err = fmt.Errorf("%s not found", name)
		return
	}
	in := make([][]byte, len(args))
	for i, it := range args {
		if fn.proto.ArgTypes[i] == "string" {
			it = `"` + it + `"`
		}
		in[i] = []byte(it)
	}
	return fn.Call(in)
}

func (p *Client) reg(prefix string, obj interface{}) (err error) {
	fns := NewFnTraits(obj)
	for origin, fn := range fns {
		name := prefix + origin
		if _, ok := p.fns[name]; ok {
			err = fmt.Errorf("%s has registered", name)
			return
		}
		fn.Prefix(prefix)
		p.fns[name] = fn
		fn.Bind(p.addr + name)
	}
	return
}

func NewClient(addr string) (p *Client) {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	if !strings.HasSuffix(addr, "/") {
		addr = addr + "/"
	}

	p = &Client{addr: addr, fns: make(FnTraits)}

	err := p.reg(MeasurePrefix, &p.Measure)
	if err != nil {
		panic(err)
	}

	err = p.reg(BuiltinPrefix, &p.Builtin)
	if err != nil {
		panic(err)
	}
	return
}

type Client struct {
	addr string
	fns FnTraits
	Measure Measure
	Builtin Builtin
}

type Measure struct {
	Sync func(time int64) (measure.MeasureData, error) `args:"time" return:"data"`
}

type Builtin struct {
	List   func() ([]FnProto, error) `return:"protos"`
	Uptime func() (start int64, dura int64, err error) `return:"start,dura"`
}

var ErrApiExists = errors.New("api registered")

const  MeasurePrefix = "Measure."
const  BuiltinPrefix = "Builtin."
