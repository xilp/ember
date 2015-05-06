package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

func (p OutValue) Interface() (ret []interface{}, err error) {
	ret = make([]interface{}, len(p))
	for i := 0; i < len(ret); i++ {
		ret[i] = p[i].Interface()
	}
	err = IsError{p[len(p) - 1].Interface()}.Check()
	ret = ret[:len(ret) - 1]
	return
}

type OutValue []reflect.Value

func (p InValue) Make(fn reflect.Value) (in []reflect.Value, err error) {
	ajust := fn.Type().NumIn() - len(p)
	in = make([]reflect.Value, len(p))
	for i, arg := range p {
		typ := fn.Type().In(i + ajust)
		val := reflect.New(typ)
		err = json.Unmarshal(arg, val.Interface())
		if err != nil {
			err = fmt.Errorf(
				"arg #%v(%v) marshal failed: '%v' => %v",
				i, typ.Kind().String(), string(arg), err.Error())
			return
		}
		in[i] = val.Elem()
	}
	return
}

type InValue [][]byte

func (p ApiFunc) Valid() (err error) {
	defer func() {
		if err != nil {
			return
		}
		err = IsError{recover()}.Check()
	}()

	fn := reflect.Value(p)

	err = ErrReturnTypeNotMatched
	out := fn.Type().NumOut()
	if out == 0 || fn.Type().Out(out - 1).Kind() != reflect.Interface {
		return
	}
	if _, is := fn.Type().Out(out - 1).MethodByName("Error"); !is {
		return
	}
	return nil
}

type ApiFunc reflect.Value

var ErrReturnTypeNotMatched = errors.New("last return value not type error")

func (p IsError) Check() (err error) {
	if p.E == nil {
		return
	}
	if r, ok := p.E.(error); ok {
		err = r
	} else if s, ok := p.E.(string); ok {
		err = errors.New(s)
	} else {
		err = ErrCallUnknown
	}
	return
}

type IsError struct{E interface{}}

var ErrCallUnknown = errors.New("unknown error type on call api")
