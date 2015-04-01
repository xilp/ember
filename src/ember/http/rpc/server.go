package rpc

import (
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type ApiTrait interface {
	Trait() map[string][]string
}

const (
	StatusOK = "OK"
	StatusErr = "error"
	HttpCodeOK = http.StatusOK
	HttpCodeErr = 599
)

var ErrUnknown = errors.New("unknown error type on call api")

type ErrRpcServer struct {
	err error
}

func NewErrRpcServer(e error) *ErrRpcServer {
	return &ErrRpcServer{e}
}

func (p *ErrRpcServer) Error() string {
	return "rpc server error: " + p.err.Error()
}

type Server struct {
	fns map[string]reflect.Value
	objs map[string]interface{}
	trait map[string][]string
}

func NewServer() *Server {
	return &Server {
		fns: make(map[string]reflect.Value),
		objs: make(map[string]interface{}),
		trait: make(map[string][]string),
	}
}

func (p *Server) Run(port int) error {
	http.HandleFunc("/", p.router)
	return http.ListenAndServe(":" + strconv.Itoa(port), nil)
}

func (p *Server) Reg(api ApiTrait) (err error) {
	typ := reflect.TypeOf(api)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		name := method.Name
		if (name == "Trait") {
			continue
		}
		err = p.register(name, api, method.Func.Interface())
		if err != nil {
			return
		}
	}
	return
}

func (p *Server) register(name string, api ApiTrait, fn interface{}) (err error) {
	fv := reflect.ValueOf(fn)

	err = callable(fv)
	if err != nil {
		return
	}

	nOut := fv.Type().NumOut()
	if nOut == 0 || fv.Type().Out(nOut - 1).Kind() != reflect.Interface {
		err = NewErrRpcServer(fmt.Errorf("%s return final output param must be error interface", name))
		return
	}

	_, b := fv.Type().Out(nOut - 1).MethodByName("Error")
	if !b {
		err = NewErrRpcServer(fmt.Errorf("%s return final output param must be error interface", name))
		return
	}

	if _, ok := p.fns[name]; ok {
		err = NewErrRpcServer(fmt.Errorf("%s has registered", name))
		return
	}

	for fn, args := range api.Trait() {
		p.trait[fn] = args
	}
	p.fns[name] = fv
	p.objs[name] = api
	return
}

func (p *Server) router(w http.ResponseWriter, r *http.Request) {
	var status string
	var detail string
	result, err := p.handle(w, r)
	if err == nil {
		status = StatusOK
	} else {
		status = StatusErr
		result = nil
		detail = err.Error()
	}

	resp := NewResponse(status, detail, result)
	ret, err := json.Marshal(resp)
	if err != nil {
		if detail != "" {
			panic("error marshal: must OK")
		}
		resp = NewResponse(StatusErr, NewErrRpcServer(err).Error(), nil)
		ret, err = json.Marshal(resp)
		if err != nil {
			panic("error marshal: must OK, too")
		}
	}

	h := w.Header()
	h.Set("Content-Type", "text/json")

	if resp.Status == StatusOK {
		w.WriteHeader(HttpCodeOK)
	} else {
		w.WriteHeader(HttpCodeErr)
	}

	_, err = w.Write(ret)
	// TODO: log here
}

func (p *Server) handle(w http.ResponseWriter, r *http.Request) (result []interface{}, err error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var in map[string]json.RawMessage
	err = json.Unmarshal(data, &in)
	if err != nil {
		return
	}

	url := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	name := url[len(url) - 1]
	return p.invoke(name, in)
}

func (p *Server) invoke(name string, args map[string]json.RawMessage) (ret []interface{}, err error) {
	fn, ok := p.fns[name]

	if !ok {
		err = NewErrRpcServer(fmt.Errorf("api %s not registered", name))
		return
	}
	if len(args) + 1 != fn.Type().NumIn() {
		err = NewErrRpcServer(fmt.Errorf("api %s params count unmatched", name))
		return
	}

	in := make([]reflect.Value, fn.Type().NumIn())
	in[0] = reflect.ValueOf(p.objs[name])

	for i, argName := range p.trait[name] {
		if args[argName] == nil {
			in[i + 1] = reflect.Zero(fn.Type().In(i))
		} else {
			typ := fn.Type().In(i + 1)
			val := reflect.New(typ)
			if _, ok := args[argName]; !ok {
				return nil, NewErrRpcServer(fmt.Errorf("arg %s missing", argName))
			}
			err = json.Unmarshal(args[argName], val.Interface())
			if err != nil {
				return nil, NewErrRpcServer(err)
			}
			in[i + 1] = val.Elem()
		}
	}

	out, err := call(fn, in)
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
			err = NewErrRpcServer(e)
		} else if e, ok := pv.(string); ok {
			err = NewErrRpcServer(fmt.Errorf(e))
		}
		return nil, err
	}

	return ret, nil
}

func NewResponse(status, detail string, result []interface{}) *Response {
	return &Response {
		status,
		detail,
		result,
	}
}

type Response struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
	Result []interface{} `json:"result"`
}

func callable(fn reflect.Value) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if r, ok := e.(error); ok {
				err = r
			} else if s, ok := e.(string); ok {
				err = NewErrRpcServer(errors.New(s))
			} else {
				err = NewErrRpcServer(ErrUnknown)
			}
		}
	}()
	fn.Type().NumIn()
	return
}

func call(fn reflect.Value, in []reflect.Value) (out []reflect.Value, err error) {
	defer func() {
		e := recover()
		if e != nil {
			if r, ok := e.(error); ok {
				err = r
			} else if s, ok := e.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.New("unknown error type on call api")
			}
		}
	}()
	out = fn.Call(in)
	return
}
