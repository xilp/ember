package main

import (
	"os"
	"strconv"
	"ember/http/rpc"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: demo port")
		os.Exit(1)
	}

	err := Launch(os.Args[1:])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func Launch(args []string) (err error) {
	port, err := strconv.Atoi(args[0])
	if err != nil {
		return
	}

	s := NewServer()
	rpc := rpc.NewServer()
	err = rpc.Register(s)
	if err != nil {
		return
	}

	return rpc.Run(port)
}
