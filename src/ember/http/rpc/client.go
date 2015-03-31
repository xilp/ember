package rpc

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

type Client struct {
	url string
	trait map[string][]string
	funs map[string]interface{}
}

func NewClient(url string) *Client {
	return &Client{
		url: url + "/",
		trait: make(map[string][]string),
		funs: make(map[string]interface{}),
	}
}

func (p *Client) MakeRpc(obj interface{}, api ApiTrait) (err error) {
	typ := reflect.TypeOf(obj).Elem()
	for i := 0; i < typ.NumField(); i++ {
		val := reflect.ValueOf(obj).Elem()
		structField := typ.Field(i)
		name := structField.Name
		field := val.Field(i)
		err = p.create(name, api, field.Addr().Interface())
		if err != nil {
			return
		}
	}
	return
}

func (p *Client) create(name string, api ApiTrait, fptr interface{}) (err error) {
	defer func() {
		e := recover()
		if e != nil {
			if r, ok := e.(error); ok {
				err = r
			} else {
				err = errors.New(e.(string))
			}
		}
	}()

	fn := reflect.ValueOf(fptr).Elem()

	nOut := fn.Type().NumOut();
	if nOut == 0 || fn.Type().Out(nOut - 1).Kind() != reflect.Interface {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	_, ok := fn.Type().Out(nOut - 1).MethodByName("Error")
	if !ok {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	for fun, args := range api.Trait() {
		p.trait[fun] = args
	}

	fun := func(in []reflect.Value) []reflect.Value {
		return p.call(fn, name, in)
	}

	fv := reflect.MakeFunc(fn.Type(), fun)
	fn.Set(fv)
	return
}

// TODO: cli
func (p *Client) Call(name string, jsonArgs string) []interface{} {
	return nil
}

func (p *Client) call(fn reflect.Value, name string, in []reflect.Value) []reflect.Value {
	nameValuePair := make(map[string]interface{})
	for i, argName := range p.trait[name] {
		nameValuePair[argName] = in[i].Interface()
	}
	inData, err := json.Marshal(nameValuePair)
	if err != nil {
		return p.returnCallError(fn, err)
	}

	resp, err := http.Post(p.url + name, "text/json", bytes.NewReader(inData))
	if err != nil {
		return p.returnCallError(fn, err)
	}

	defer resp.Body.Close()
	outData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.returnCallError(fn, err)
	}

	var outJson struct {
		Status string
		Detail string
		Result []json.RawMessage
	}

	err = json.Unmarshal(outData, &outJson)
	if err != nil {
		return p.returnCallError(fn, err)
	}

	if outJson.Status != StatusOK {
		return p.returnCallError(fn, errors.New(outJson.Detail))
	}

	out := make([]reflect.Value, fn.Type().NumOut())
	for i := 0; i < len(out); i++ {
		if len(outJson.Result) <= i || outJson.Result[i] == nil {
			out[i] = reflect.Zero(fn.Type().Out(i))
		} else {
			typ := fn.Type().Out(i)
			val := reflect.New(typ)
			err = json.Unmarshal(outJson.Result[i], val.Interface())
			if err != nil {
				return p.returnCallError(fn, err)
			}
			out[i] = val.Elem()
		}
	}

	return out
}

func (c *Client) returnCallError(fn reflect.Value, err error) []reflect.Value {
	nOut := fn.Type().NumOut()
	out := make([]reflect.Value, nOut)
	for i := 0; i < nOut - 1; i++ {
		out[i] = reflect.Zero(fn.Type().Out(i))
	}

	out[nOut-1] = reflect.ValueOf(&err).Elem()
	return out
}

func NewInArgs(args map[string]interface{}) *InArgs {
	return &InArgs{args}
}

type InArgs struct {
	//Args interface{} `json:"args"`
	Args map[string]interface{} `json:"args"`

}
