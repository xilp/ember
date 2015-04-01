package rpc

import (
	"bytes"
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"ember/measure"
)

type Client struct {
	url string
	trait map[string][]string
	fns map[string]interface{}
}

func NewClient(url string) (p *Client) {
	p = &Client {
		url: url + "/",
		trait: make(map[string][]string),
		fns: make(map[string]interface{}),
	}

	var m struct {
		MeasureSync func(time int64) (measure.MeasureData, error)
	}
	err := p.reg(&m, MeasureTrait)
	if err != nil {
		panic(err)
	}
	return
}

func (p *Client) Reg(obj interface{}, api ApiTrait) (err error) {
	return p.reg(obj, api.Trait())
}

func (p *Client) reg(obj interface{}, trait map[string][]string) (err error) {
	typ := reflect.TypeOf(obj).Elem()
	for i := 0; i < typ.NumField(); i++ {
		val := reflect.ValueOf(obj).Elem()
		structField := typ.Field(i)
		name := structField.Name
		field := val.Field(i)
		if callable(field) != nil {
			continue
		}
		err = p.create(name, trait, field.Addr().Interface())
		if err != nil {
			return
		}
	}
	return
}

func (p *Client) create(name string, trait map[string][]string, fptr interface{}) (err error) {
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

	for fn, args := range trait {
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

func (p *Client) List() (ret []string) {
	for k, _ := range p.fns {
		ret = append(ret, k)
	}
	return
}

func (p *Client) Invoke(args []string) (ret []interface{}, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("missed api name. all: %v", p.List())
		return
	}

	name := args[0]
	args = args[1:]
	fn := p.fns[name]
	if fn == nil {
		err = fmt.Errorf("'%s' not found. all: %v", name, p.List())
		return
	}

	fv := reflect.ValueOf(fn)

	nOut := fv.Type().NumOut() - 1
	if nOut != len(args) || len(p.trait[name]) != len(args) {
		err = fmt.Errorf("'%s' args list %v unmatched (need %d, got %d)", name, p.trait[name], len(args), nOut)
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

func (p *Client) Call(args []string) (ret string, err error) {
	objs, err := p.Invoke(args)
	if err != nil {
		return
	}

	for i := 0; i < len(objs) - 1; i++ {
		val := fmt.Sprintf("%#v", objs[i])
		if val[0] == '"' && val[len(val) - 1] =='"' && len(val) > 2 {
			val = val[1:len(val) - 1]
		}
		ret += val
		if i + 1 != len(objs) - 1 {
			ret += ", "
		}
	}
	return
}
