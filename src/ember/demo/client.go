package main

import (
	"ember/http/rpc"
)

func NewClient(addr string) (p *Client, err error) {
	p = &Client{Rpc: rpc.NewClient(addr)}
	err = p.Rpc.Reg(p, &Server{})
	return
}

type Client struct {
	Rpc *rpc.Client
	Echo func(msg string) (echo string, err error)
	Panic func() (err error)
	Stop func() (err error)
}
