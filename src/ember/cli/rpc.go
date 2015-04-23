package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"ember/http/rpc"
)

func (p *RpcHub) Run() {
	if len(p.args) == 0 {
		Errln("usage:\n  <bin> [-host=" + DefaultHost + "] [-port=" + DefaultPort + "] command [args]\n\ncommand:")
		p.cmds.Help(true)
		os.Exit(1)
	}
	p.cmds.Run(p.args)
}

func (p *RpcHub) Cmds() *Cmds {
	return p.cmds
}

func (p *RpcHub) CmdRun([]string) {
	err := p.server.Run("/", p.port)
	Check(err)
}

func (p *RpcHub) CmdList([]string) {
	fns := p.client.List()
	p.help(fns)
}

func (p *RpcHub) CmdRemote([]string) {
	fns, err := p.client.Builtin.List()
	Check(err)
	p.help(fns)
}

func (p *RpcHub) help(fns []rpc.FnProto) {
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
		fmt.Printf("%s%v => %v\n",
			fn.Name,
			types(fn.ArgNames, fn.ArgTypes, "(", ")"),
			types(fn.ReturnNames, fn.ReturnTypes, "(", ")"))
	}
}

func (p *RpcHub) CmdCall(args []string) {
	if len(args) == 0 {
		p.CmdList(args)
		return
	}
	ret, err := p.client.Call(args[0], args[1:])
	Check(err)

	var obj interface{}
	obj = ret
	if len(ret) == 1 {
		obj = ret[0]
	}

	data, err := json.MarshalIndent(obj, "", "    ")
	Check(err)
	fmt.Println(string(data))
}

func (p *RpcHub) CmdStatus(args []string) {
	data, err := p.client.Measure.Sync(0)
	err = data.Dump(os.Stdout, true)
	Check(err)
}

func NewRpcHub(args []string, sobj interface{}, cobj interface{}) (p *RpcHub)  {
	host, args := PopArg("host", DefaultHost, args)
	portstr, args := PopArg("port", DefaultPort, args)
	port, err := strconv.Atoi(portstr)
	Check(err)

	addr := host + ":" + portstr

	client := rpc.NewClient(addr)
	err = client.Reg(cobj)
	Check(err)

	server := rpc.NewServer()
	err = server.Reg(sobj, cobj)
	Check(err)

	p = &RpcHub{host, port, args, NewCmds(), server, client}

	p.cmds.Reg("run", "run server", p.CmdRun)
	p.cmds.Reg("list", "list api from local info", p.CmdList)
	p.cmds.Reg("remote", "list api from remote", p.CmdRemote)
	p.cmds.Reg("call", "call server api by: name [arg] [arg]...", p.CmdCall)
	p.cmds.Reg("status", "get server status", p.CmdStatus)
	return
}

type RpcHub struct {
	host string
	port int
	args []string
	cmds *Cmds
	server *rpc.Server
	client *rpc.Client
}

const DefaultHost = "127.0.0.1"
const DefaultPort = "8080"
