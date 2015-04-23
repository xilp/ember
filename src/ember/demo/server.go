package main

import (
	"errors"
	"os"
	"ember/cli"
)

func main() {
	hub := cli.NewRpcHub(os.Args[1:], &Server{}, &Client{})
	hub.Run()
}

type Client struct {
	Echo func(msg string) (echo string, err error) `args:"msg" return:"echo"`
	Panic func() (err error)
	Error func() (err error)
	Foo func() (ret [][][]string, err error) `return:"ret"`
}

func (p *Server) Echo(msg string) (echo string, err error) {
	echo = msg
	return
}

func (p *Server) Panic() (err error) {
	panic("panic as expected")
	return
}

func (p *Server) Error() (err error) {
	err = errors.New("error as expected")
	return
}

func (p *Server) Foo() (ret [][][]string, err error) {
	ret = [][][]string{{{"foo"}}}
	return
}

type Server struct{}
