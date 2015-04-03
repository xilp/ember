package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"ember/cli"
	"ember/measure"
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
	cmds.Reg("status", "get server core status", client.CmdStatus)

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
	fmt.Println(ret)
}

func (p *Client) CmdStatus(args []string) {
	ret, err := p.Rpc.Invoke([]string{"MeasureSync", "0"})
	cli.Check(err)
	if len(ret) != 2 {
		fmt.Println(ret)
	} else {
		data := ret[0].(measure.MeasureData)
		err = data.Dump(os.Stdout, true)
		cli.Check(err)
	}
}

func (p *Client) CmdStop(args []string) {
	p.CmdCall([]string{"Stop"})
}
