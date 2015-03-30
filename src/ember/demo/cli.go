package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"ember/cli"
)

const ServerPort = 8080

func main() {
	client, err := NewClient("http://127.0.0.1:" + strconv.Itoa(ServerPort))
	cli.Check(err)

	cmds := cli.NewCmds()
	cmds.Reg("run", "run server", CmdRun)
	cmds.Reg("stop", "stop server", client.CmdStop)

	api:= cmds.Sub("api", "call server api on 127.0.0.1")
	api.Reg("echo", "echo message", client.CmdEcho)
	api.Reg("panic", "server panic test", client.CmdPanic)
	api.Reg("stop", "stop server", client.CmdStop)

	cmds.Run(os.Args[1:])
}

func CmdRun(args []string) {
	err := Launch(ServerPort)
	cli.Check(err)
}

func (p *Client) CmdEcho(args []string) {
	flag := flag.NewFlagSet("echo", flag.PanicOnError)
	msg := flag.String("msg", "", "message to be sent to server")
	cli.ParseFlag(flag, args, "msg")

	echo, err := p.Echo(*msg)
	cli.Check(err)
	fmt.Println(echo)
}

func (p *Client) CmdPanic(args []string) {
	cli.ParseFlag(flag.NewFlagSet("panic", flag.PanicOnError), args)
	cli.Check(p.Panic())
}

func (p *Client) CmdStop(args []string) {
	cli.ParseFlag(flag.NewFlagSet("panic", flag.PanicOnError), args)
	cli.Check(p.Stop())
}
