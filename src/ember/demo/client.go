package main

import (
	"ember/http/rpc"
)

func NewClient(addr string) (p *Client, err error) {
	p = &Client{}
	err = rpc.NewClient(addr).MakeRpc(p)
	return
}

type Client struct {
	Echo func(msg string) (echo string, err error)
	Panic func() (err error)
	Stop func() (err error)
}
