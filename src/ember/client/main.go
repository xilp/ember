package main

import (
	"fmt"
	"ember/http/rpc"
	"ember/cli"
	"encoding/json"
	"os"
)

func main() {
	args := os.Args[1:]
	c := NewEmberClient(args, "/")
	if len(args) > 1 {
		c.Call(args)
	} else {
		c.List(args)
	}
}

func (p *EmberClient) List([]string) {
	fns, err := p.client.Builtin.List()
	cli.Check(err)
	p.help(fns)
}

func (p *EmberClient) Call(args []string) {
	if len(args) == 0 {
		return
	}
	ret, err := p.client.Call(args[0], args[1:])
	cli.Check(err)

	if len(ret) == 0 || ret == nil {
		return
	}

	var obj interface{}
	obj = ret
	if len(ret) == 1 {
		obj = ret[0]
	}

	data, err := json.MarshalIndent(obj, "", "    ")
	cli.Check(err)
	fmt.Println(string(data))
}

func (p *EmberClient) help(fns []rpc.FnProto) {
	types := func(names []string, types []string, lb, rb string) string {
		str := lb
		for i, name := range names {
			str += types[i] + " " + name
			if i + 1 != len(names) {
				str += ", "
			}
		}
		return str + rb
	}

	for _, fn := range fns {
		fmt.Printf("  %s%v => %v\n",
			fn.Name,
			types(fn.ArgNames, fn.ArgTypes, "(", ")"),
			types(fn.ReturnNames, fn.ReturnTypes, "(", ")"))
	}
}

func NewEmberClient(args []string, path string) (p *EmberClient) {
	host, args := cli.PopArg("host", "127.0.0.1", args)
	portStr, args := cli.PopArg("port", "8088", args)

	addr := host + ":" + portStr + path

	client := rpc.NewClient(addr)
	p = &EmberClient{client, path, args}
	return
}

type EmberClient struct {
	client *rpc.Client
	path string
	args []string
}
