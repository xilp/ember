package main

import (
	"errors"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (p *Server) Echo(msg string) (echo string, err error) {
	echo = msg
	return
}

func (p *Server) Panic() (err error) {
	return errors.New("panic as expected")
}
