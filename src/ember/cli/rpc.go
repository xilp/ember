package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"ember/http/rpc"
)

func (p *RpcHub) Run() {
	if len(p.args) == 0 {
		Errln("usage:\n  <bin> [-host=127.0.0.1] [-port=8080] command [args]\n\ncommand:")
		p.cmds.Help(true)
		os.Exit(1)
	}
	p.cmds.Run(p.args)
}

func (p *RpcHub) Cmds() *Cmds {
	return p.cmds
}

func (p *RpcHub) CmdRun([]string) {
	err := p.server.Run(p.port)
	Check(err)
}

func (p *RpcHub) CmdList([]string) {
	fns, err := p.client.Builtin.List()
	Check(err)

	apis := []string{}
	builtins := []string{}
	for name, _ := range fns {
		if strings.Index(name, ".") < 0 {
			apis = append(apis, name)
		} else {
			builtins = append(builtins, name)
		}
	}
	sort.Strings(apis)
	sort.Strings(builtins)

	display := func(names []string) {
		for _, name := range names {
			args := fns[name]
			if len(args) == 0 {
				fmt.Println(name + " []")
			} else {
				fmt.Printf("%s %v\n", name, args)
			}
		}
	}

	display(apis)
	display(builtins)
}

func (p *RpcHub) CmdCall(args []string) {
	if len(args) == 0 {
		p.CmdList(args)
		return
	}
	ret, err := p.client.Invoke(args)
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

func NewRpcHub(args []string, sobj rpc.ApiTrait, cobj interface{}) (p *RpcHub)  {
	host, args := PopArg("host", "127.0.0.1", args)
	portstr, args := PopArg("port", "8080", args)
	port, err := strconv.Atoi(portstr)
	Check(err)

	addr := host + ":" + portstr
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	client := rpc.NewClient(addr)
	err = client.Reg(cobj, sobj)
	Check(err)

	server := rpc.NewServer()
	err = server.Reg(sobj)
	Check(err)

	p = &RpcHub{host, port, args, NewCmds(), server, client}

	p.cmds.Reg("run", "run server", p.CmdRun)
	p.cmds.Reg("list", "list server api", p.CmdList)
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
