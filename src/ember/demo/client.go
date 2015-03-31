package main

import (
	"ember/http/rpc"
)

func NewClient(addr string) (p *Client, err error) {
	p = &Client{}
	c := rpc.NewClient(addr)
	err = c.MakeRpc(p, &Server{})
	return
}

type Client struct {
	Echo func(msg string) (echo string, err error)
	Panic func() (err error)
	Stop func() (err error)
}
