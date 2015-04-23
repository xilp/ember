package rpc

import (
	"reflect"
	"strings"
)

func NewFnProto(fn reflect.Value, tag reflect.StructTag, name string) (p FnProto) {
	p = FnProto{Name: name}
	typ := fn.Elem().Type()
	for i := 0; i < typ.NumIn(); i++ {
		tn := typ.In(i).Name()
		if tn == "" {
			tn = typ.In(i).Kind().String()
		}
		p.ArgTypes = append(p.ArgTypes, tn)
	}
	for i := 0; i < typ.NumOut(); i++ {
		tn := typ.Out(i).Name()
		if tn == "" {
			tn = typ.Out(i).Kind().String()
		}
		p.ReturnTypes = append(p.ReturnTypes, tn)
	}

	args := tag.Get("args")
	if len(args) == 0 {
		p.ArgNames = []string{}
	} else {
		for _, it := range strings.Split(args, ",") {
			p.ArgNames = append(p.ArgNames, strings.TrimSpace(it))
		}
	}

	ret := tag.Get("return")
	if len(ret) == 0 {
		p.ReturnNames = []string{}
	} else {
		for _, it := range strings.Split(ret, ",") {
			p.ReturnNames = append(p.ReturnNames, strings.TrimSpace(it))
		}
	}
	return
}

type FnProto struct {
	Name        string   "name"
	ArgNames    []string "arg_names"
	ArgTypes    []string "arg_types"
	ReturnNames []string "return_names"
	ReturnTypes []string "return_types"
}
