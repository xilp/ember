package rpc

import (
	"reflect"
	"sync"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"bytes"
	"time"
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
	for i := 0; i < typ.NumField(); i++ {
		val := reflect.ValueOf(obj).Elem()
		structField := typ.Field(i)
		name := structField.Name
		field := val.Field(i)
		err = c.makeRpc(name, field.Addr().Interface())
	}
	return
}

func (c *Client) makeRpc(rpcName string, fptr interface{}) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if _, ok := e.(error); ok {
				err = fmt.Errorf("make rpc error", e.(error).Error())
			} else {
				err = fmt.Errorf("make rpc error", e.(string))
			}
		}
	}()

	fn := reflect.ValueOf(fptr).Elem()

	nOut := fn.Type().NumOut();
	if nOut == 0 || fn.Type().Out(nOut - 1).Kind() != reflect.Interface {
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

	input := NewInArgs(inArgs)
	data, err := json.Marshal(input)

	resp, err := http.Post(c.url + name, "text/json", bytes.NewReader(data))
	if err != nil {
		println(err.Error())
	}

	dataBytes, err := ioutil.ReadAll(resp.Body)

	time.Sleep(time.Second * 3)
	if err != nil {
		return c.returnCallError(fn, err)
	}

	var f struct {
		Result []json.RawMessage
	}

	errs := json.Unmarshal(dataBytes, &f)
	if errs != nil {
		return c.returnCallError(fn, errs)
	}

	outValues := make([]reflect.Value, len(f.Result))
	for i := 0; i < len(f.Result); i++ {
		if f.Result[i] == nil {
			outValues[i] = reflect.Zero(fn.Type().Out(i))
		} else {
			typ := fn.Type().Out(i)
			newVal := reflect.New(typ)
			err := json.Unmarshal(f.Result[i], newVal.Interface())
			if err != nil {
				panic("FXXX")
			}
			outValues[i] = newVal.Elem()

		}
	}
	defer resp.Body.Close()
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

func NewInArgs(args []interface{}) *InArgs {
	a := new(InArgs)
	a.Args = args
	return a
}

type InArgs struct {
	Args []interface{}
}
