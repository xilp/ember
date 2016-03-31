package rpc

import (
	"encoding/json"
	"errors"
	"reflect"
)

func NewFnValues(obj interface{}, traits FnTraits) (fvs FnValues, err error) {
	fvs = FnValues{}
	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if _, ok := traits[method.Name]; !ok {
			continue
		}
		fn := reflect.ValueOf(method.Func.Interface())
		if ApiFunc(fn).Valid() != nil {
			continue
		}
		fvs[method.Name] = FnValue{val, fn}
	}
	return
}

type FnValues map[string]FnValue

func (p FnValue) Invoke(proto FnProto, data []byte) (ret []interface{}, err error) {
	var in map[string]json.RawMessage
	if len(data) != 0 {
		err = json.Unmarshal(data, &in)
		if err != nil {
			return
		}
	}
	args := make([][]byte, len(proto.ArgNames))
	for i, name := range proto.ArgNames {
		arg, ok := in[name]
		if !ok {
			err = ErrArgsNotMatched
			return
		}
		args[i] = arg
	}

	return p.Call(args)
}

func (p FnValue) Call(args [][]byte) (ret []interface{}, err error) {
	defer func() {
		if err != nil {
			return
		}
		err = IsError{recover()}.Check()
	}()

	in, err := InValue(args).Make(p.fn)
	if err != nil {
		return
	}
	in = append([]reflect.Value{p.this}, in...)

	out := p.fn.Call(in)
	ret, err = OutValue(out).Interface()
	if err != nil {
		err = NewErrCallFailed(err)
	}
	return
}

type FnValue struct {
	this  reflect.Value
	fn    reflect.Value
}

func (p *ErrCallFailed) Error() string {
	return "[invoke failed] " + p.err.Error()
}

func NewErrCallFailed(err error) *ErrCallFailed {
	return &ErrCallFailed{err}
}

type ErrCallFailed struct {
	err error
}

var ErrArgsNotMatched = errors.New("args not matched")
