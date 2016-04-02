package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

func (p FnTraits) List() (fns []FnProto) {
	builtin := []string{}
	api := []string{}
	for name, _ := range p {
		if strings.Index(name, ".") >= 0 {
			builtin = append(builtin, name)
		} else {
			api = append(api, name)
		}
	}
	sort.Strings(builtin)
	sort.Strings(api)

	for _, name := range builtin {
		fns = append(fns, p[name].Proto())
	}
	for _, name := range api {
		fns = append(fns, p[name].Proto())
	}
	return
}

func NewFnTraits(obj interface{}) (fns FnTraits) {
	fns = FnTraits{}
	typ := reflect.TypeOf(obj).Elem()
	val := reflect.Indirect(reflect.ValueOf(obj))

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fn := val.Field(i)
		if ApiFunc(fn).Valid() != nil {
			continue
		}
		fns[field.Name] = NewFnTrait(fn.Addr(), field.Tag, field.Name)
	}
	return
}

type FnTraits map[string]*FnTrait

func (p *FnTrait) Bind(addr string) {
	p.addr = addr
	fn := p.fn.Elem()
	fv := reflect.MakeFunc(fn.Type(), p.proxy)
	fn.Set(fv)
	return
}

func (p *FnTrait) Call(args [][]byte) (ret []interface{}, err error) {
	in, err := InValue(args).Make(p.fn.Elem())
	if err != nil {
		return
	}
	out := p.proxy(in)
	ret, err = OutValue(out).Interface()
	return
}

func (p *FnTrait) proxy(in []reflect.Value) (out []reflect.Value) {
	fn := p.fn.Elem()

	out = make([]reflect.Value, fn.Type().NumOut())
	var err error
	defer func() {
		if err != nil {
			for i := 0; i < len(out) - 1; i++ {
				out[i] = reflect.Zero(fn.Type().Out(i))
			}
			out[len(out) - 1] = reflect.ValueOf(&err).Elem()
		}
	}()

	if len(p.proto.ArgNames) != len(in) {
		err = ErrArgsNotMatched
		return
	}

	args := make(map[string]interface{})
	for i, name := range p.proto.ArgNames {
		args[name] = in[i].Interface()
	}

	data, err := json.Marshal(args)
	if err != nil {
		return
	}

	resp, err := http.Post(p.addr, "text/json", bytes.NewReader(data))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var ret struct {
		Status string `json:"status"`
		Detail string `json:"detail"`
		Result map[string]json.RawMessage `json:"result"`
	}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		err = fmt.Errorf("<json decode error: %v>\n%v", err.Error(), string(body))
		return
	}

	retArr := make([]json.RawMessage, len(ret.Result))
	retArgs := p.proto.ReturnNames
	for index, retArg := range retArgs {
		if (index < len(retArr)) {
			retArr[index] = ret.Result[retArg]
		}
	}

	if ret.Status != StatusOK {
		err = errors.New(ret.Detail)
		return
	}

	for i := 0; i < len(out); i++ {
		if len(retArr) <= i || retArr[i] == nil {
			out[i] = reflect.Zero(fn.Type().Out(i))
		} else {
			typ := fn.Type().Out(i)
			val := reflect.New(typ)
			err = json.Unmarshal(retArr[i], val.Interface())
			if err != nil {
				return
			}
			out[i] = val.Elem()
		}
	}
	return
}

func (p *FnTrait) Proto() (proto FnProto) {
	return p.proto
}

func (p *FnTrait) Prefix(prefix string) {
	p.proto.Name = prefix + p.proto.Name
}

func NewFnTrait(fn reflect.Value, tag reflect.StructTag, name string) (p *FnTrait) {
	return &FnTrait{proto: NewFnProto(fn, tag, name), fn: fn}
}

type FnTrait struct {
	proto FnProto
	fn    reflect.Value
	addr  string
}
