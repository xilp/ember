package main

import (
	"errors"
	"os"
	"time"
	"ember/http/rpc"
)

func Launch(port int) (err error) {
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

func (p *Server) Echo(msg string) (echo string, err error) {
	echo = msg
	return
}

func (p *Server) Panic() (err error) {
	err = errors.New("panic as expected")
	return
}

func (p *Server) Stop() (err error) {
	go func() {
		time.Sleep(time.Second)
		os.Exit(0)
	}()
	return
}

func (p *Server) Trait() map[string][]string {
	return map[string][]string {
		"Echo": []string{"msg"},
		"Panic": []string{},
		"Stop": []string{},
	}
}

func NewServer() *Server {
	return &Server{}
}

type Server struct {
}
