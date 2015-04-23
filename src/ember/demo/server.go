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
	Echo func(msg string) (echo string, err error)
	Panic func() (err error)
	Error func() (err error)
}

func (p *Server) Trait() map[string][]string {
	return map[string][]string {
		"Echo": {"msg"},
		"Panic": {},
		"Error": {},
	}
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

type Server struct{}
