package main

import (
	"errors"
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
	if len(args) != 1 {
		cli.Check(errors.New("echo msg: unmatched args count"))
	}
	msg, err := p.Echo(args[0])
	cli.Check(err)
	fmt.Println(msg)
}

func (p *Client) CmdPanic(args []string) {
	cli.Check(p.Panic())
}

func (p *Client) CmdStop(args []string) {
	cli.Check(p.Stop())
}
