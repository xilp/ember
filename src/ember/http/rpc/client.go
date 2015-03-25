package rpc

import (
	"reflect"
	"sync"
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"io/ioutil"
	"bytes"
)

type Client struct {
	sync.Mutex
	
	url string
}

func NewClient(url string) *Client {
	c := new(Client)
	c.url = url
	
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

	input := NewInArgs(name, inArgs)
	data, err := json.Marshal(in)
	
	buf, err := bytes.NewReader(data).Read(data)
	resp, err := http.Post(url + name, "text/json", &buf)
	if err != nil {
		return c.returnCallError(fn, err)
	}
	
	dataBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.returnCallError(fn, err)
	}
	
	var f interface{}
	err = json.Unmarshal(dataBytes, &f)
	if err != nil {
		return c.returnCallError(fn, err)
	}

	dataMap := f.(map[string]interface{})

	if out, ok := dataMap["args"]; !ok {
		return c.returnCallError(fn, fmt.Errorf("rpc args is nil"))
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

	defer resp.Body.Close()
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

func NewInArgs(name string, args []interface{}) *InArgs {
	a := new(InArgs)
	a.Name = name
	a.args = args
	return a
}

type InArgs struct {
    	Args []interface{} 	`json:"args"`
}


