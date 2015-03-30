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
	host, args := cli.PopArg("host", "127.0.0.1", args)
	port, args := cli.PopArg("port", "8080", args)

	client, err := NewClient("http://" + host + ":" + port)
	cli.Check(err)

	cmds := cli.NewCmds()
	cmds.Reg("run", "run server", func(args []string) { CmdRun(port, args) })
	cmds.Reg("stop", "stop server", client.CmdStop)
	cmds.Reg("call", "call server api by: api <args in json>", CmdCall)

	api:= cmds.Sub("api", "call server api")
	api.Reg("echo", "echo message", client.CmdEcho)
	api.Reg("panic", "server panic test", client.CmdPanic)
	api.Reg("stop", "stop server", client.CmdStop)

	cmds.Run(args)
}

func CmdRun(port string, args []string) {
	cli.ParseFlag(flag.NewFlagSet("run", flag.ContinueOnError), args)
	n, err := strconv.Atoi(port)
	cli.Check(err)
	err = Launch(n)
	cli.Check(err)
}

func CmdCall(args []string) {
}

func (p *Client) CmdEcho(args []string) {
	flag := flag.NewFlagSet("echo", flag.ContinueOnError)
	msg := flag.String("msg", "", "message to be sent to server")
	cli.ParseFlag(flag, args, "msg")

	echo, err := p.Echo(*msg)
	cli.Check(err)
	fmt.Println(echo)
}

func (p *Client) CmdPanic(args []string) {
	cli.ParseFlag(flag.NewFlagSet("panic test", flag.ContinueOnError), args)
	cli.Check(p.Panic())
}

func (p *Client) CmdStop(args []string) {
	cli.ParseFlag(flag.NewFlagSet("stop", flag.ContinueOnError), args)
	cli.Check(p.Stop())
}
