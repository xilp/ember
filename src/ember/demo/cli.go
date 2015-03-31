package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"ember/cli"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		cli.Errln("usage: [-host=] [-port=] cmd(eg: help)")
		os.Exit(1)
	}

	host, args := cli.PopArg("host", "127.0.0.1", args)
	port, args := cli.PopArg("port", "8080", args)

	server := &CliServer{port}
	client, err := NewClient("http://" + host + ":" + port)
	cli.Check(err)

	cmds := cli.NewCmds()
	cmds.Reg("run", "run server", server.CmdRun)
	cmds.Reg("stop", "stop server", client.CmdStop)
	cmds.Reg("api", "call server api by: name [arg] [arg]...", client.CmdCall)

	cmds.Run(args)
}

type CliServer struct {
	port string
}

func (p *CliServer) CmdRun(args []string) {
	cli.ParseFlag(flag.NewFlagSet("run", flag.ContinueOnError), args)
	n, err := strconv.Atoi(p.port)
	cli.Check(err)
	err = Launch(n)
	cli.Check(err)
}

func (p *Client) CmdCall(args []string) {
	ret, err := p.Rpc.Call(args)
	cli.Check(err)
	for i := 0; i < len(ret) - 1; i++ {
		val := fmt.Sprintf("%#v", ret[i])
		if val[0] == '"' && val[len(val) - 1] =='"' && len(val) > 2 {
			val = val[1:len(val) - 1]
		}
		fmt.Print(val)
		if i + 1 != len(ret) - 1 {
			fmt.Printf(", ")
		}
	}
	fmt.Printf("\n")
}

func (p *Client) CmdStop(args []string) {
	p.CmdCall([]string{"Stop"})
}
