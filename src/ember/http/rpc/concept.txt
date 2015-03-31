package rpc

type Server struct {
}

func (p *Server) Reg(obj interface{}) {
	// for: p.data[obj.Method.Name] = obj.Method
}

func (p *Server) Run() {
	// http.Handle(Handle)
}

type Api struct {
}

func (p *Api) Add() int {
	return 10
}

func ServerMain() {
	a := &Api{}
	s := Server{}
	s.Reg(a)
	s.Run()
}

func (p *Server) Handle() {
	// name := req.Path
	// json := req.Body
	// for k, v := json { args[k] := Unmarsh(v) }
	// call(p.data[name], args)
}

type Client struct {
}

func NewClient(url string) *Client {
	return nil
}

func (p *Client) Reg(obj interface{}) {
	// for name, method := obj { obj.method = p.Call(name) }
}

func ClientMain() {
	c := NewClient("localhost")
	var o struct {
		Add func() int
	}
	c.Reg(&o)
	println(o.Add())
}
