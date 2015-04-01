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
	fns map[string]interface{}
}

func NewClient(url string) *Client {
	return &Client{
		url: url + "/",
		trait: make(map[string][]string),
		fns: make(map[string]interface{}),
	}
}

func (p *Client) Reg(obj interface{}, api ApiTrait) (err error) {
	typ := reflect.TypeOf(obj).Elem()
	for i := 0; i < typ.NumField(); i++ {
		val := reflect.ValueOf(obj).Elem()
		structField := typ.Field(i)
		name := structField.Name
		field := val.Field(i)
		if callable(field) != nil {
			continue
		}
		err = p.create(name, api, field.Addr().Interface())
		if err != nil {
			return
		}
	}
	return
}

func (p *Client) create(name string, api ApiTrait, fptr interface{}) (err error) {
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

	for fn, args := range api.Trait() {
		p.trait[fn] = args
	}

	wrapper := func(in []reflect.Value) []reflect.Value {
		return p.invoke(fn, name, in)
	}

	fv := reflect.MakeFunc(fn.Type(), wrapper)
	fn.Set(fv)
	p.fns[name] = fn.Interface()
	return
}

func (p *Client) invoke(fn reflect.Value, name string, in []reflect.Value) []reflect.Value {
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
	Args map[string]interface{} `json:"args"`
}

func (p *Client) Call(args []string) (ret []interface{}, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("missing api name")
		return
	}

	name := args[0]
	args = args[1:]
	fn := p.fns[name]
	if fn == nil {
		err = fmt.Errorf("api %s not found", name)
		return
	}

	fv := reflect.ValueOf(fn)

	if fv.Type().NumOut() - 1 != len(args) || len(p.trait[name]) != len(args) {
		err = fmt.Errorf("api %s params count unmatched(%d/%d)", name, len(args), fv.Type().NumOut() - 1)
		return
	}

	in := make([]reflect.Value, len(args))

	for i, arg := range args {
		if arg == "" {
			in[i] = reflect.Zero(fv.Type().In(i))
		} else {
			typ := fv.Type().In(i)
			val := reflect.New(typ)
			if typ.Kind() == reflect.String {
				arg = "\"" + arg + "\""
			}
			err = json.Unmarshal([]byte(arg), val.Interface())
			if err != nil {
				return nil, err
			}
			in[i] = val.Elem()
		}
	}

	out, err := call(fv, in)
	if err != nil {
		return nil, err
	}

	ret = make([]interface{}, len(out))
	for i := 0; i < len(ret); i++ {
		ret[i] = out[i].Interface()
	}

	pv := out[len(out) - 1].Interface()
	if pv != nil {
		if e, ok := pv.(error); ok {
			err = e
		} else if e, ok := pv.(string); ok {
			err = fmt.Errorf(e)
		}
		return nil, err
	}

	return ret, nil
}
