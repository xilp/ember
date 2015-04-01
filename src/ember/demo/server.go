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
	err = rpc.Reg(s)
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
	panic("panic as expected")
	return
}

func (p *Server) Error() (err error) {
	err = errors.New("error as expected")
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
		"Stop": []string{},
		"Echo": []string{"msg"},
		"Panic": []string{},
		"Error": []string{},
	}
}

func NewServer() *Server {
	return &Server{}
}

type Server struct {
}
