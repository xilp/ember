package rpc

import (
	"reflect"
	"sync"
)

type Client struct {
	sync.Mutex
	
	network string
	addr string
}

func NewClient(network, addr string) *Client {
	RegisterType(RpcError{})

	c := new(Client)
	c.network = network
	c.addr = addr
	
	return c
}

func (c *Client) MakeRpcObj(obj interface{}) (err error) {
	typ := reflect.TypeOf(obj).Elem()
	val := reflect.ValueOf(obj).Elem()
	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		name := structField.Name
		field := val.Field(i)
		err = c.makeRpc(name, field.Addr().Interface())
		return
	}
	return
}

func (c *Client) makeRpc(rpcName string, fptr interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("make rpc error", e.(error).Error())
		}
	}()

	fn := reflect.ValueOf(fptr).Elem()

	nOut := fn.Type().NumOut();
	if nOut == 0 || fn.Type().Out(nOut-1).Kind() != reflect.Interface {
		err = fmt.Errorf("%s return final output param must be error interface", rpcName)
		return
	}

	_, b := fn.Type().Out(nOut - 1).MethodByName("Error")
	if !b {
		err = fmt.Errorf("%s return final output param must be error interface", rpcName)
		return
	}

	f := func(in []reflect.Value) []reflect.Value {
		return c.call(fn, rpcName, in)
	}

	v := reflect.MakeFunc(fn.Type(), f)
	fn.Set(v)
	return 
}

func (c *Client) call(fn reflect.Value, name string, in []reflect.Value) []reflect.Value {
	inArgs := make([]interface{}, len(in))
	for i := 0; i < len(in); i++ {
		inArgs[i] = in[i].Interface()
	}

	data, err := encodeData(name, inArgs)
	if err != nil {
		return c.returnCallError(fn, err)
	}

	var co *conn
	var buf []byte
	for i := 0; i < 3; i++ {
		if co, err = c.popConn(); err != nil {
			continue
		}

		buf, err = co.Call(data)
		if err == nil {
			c.pushConn(co)
			break
		} else {
			co.Close()
		}
	}

	if err != nil {
		return c.returnCallError(fn, err)
	}

	n, out, e := decodeData(buf)
	if e != nil {
		return c.returnCallError(fn, e)
	}

	if n != name {
		return c.returnCallError(fn, fmt.Errorf("rpc name %s != %s", n, name))
	}

	last := out[len(out)-1]
	if last != nil {
		if err, ok := last.(error); ok {
			return c.returnCallError(fn, err)
		} else {
			return c.returnCallError(fn, fmt.Errorf("rpc final return type %T must be error", last))
		}
	}

	outValues := make([]reflect.Value, len(out))
	for i := 0; i < len(out); i++ {
		if out[i] == nil {
			outValues[i] = reflect.Zero(fn.Type().Out(i))
		} else {
			outValues[i] = reflect.ValueOf(out[i])
		}
	}

	return outValues
}

func (c *Client) returnCallError(fn reflect.Value, err error) []reflect.Value {
	nOut := fn.Type().NumOut()
	out := make([]reflect.Value, nOut)
	for i := 0; i < nOut-1; i++ {
		out[i] = reflect.Zero(fn.Type().Out(i))
	}

	out[nOut-1] = reflect.ValueOf(&err).Elem()
	return out
}





















