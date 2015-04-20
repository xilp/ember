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
	"time"
	"ember/measure"
)

var ErrUnknown = errors.New("unknown error type on call api")

func (p *Server) List() (apis map[string][]string, err error) {
	apis = p.trait
	return
}

func (p *Server) Run(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.Serve)
	return http.ListenAndServe(":" + strconv.Itoa(port), mux)
}

func (p *Server) Reg(api ApiTrait) (err error) {
	return p.reg("", api, api.Trait())
}

func (p *Server) reg(prefix string, api interface{}, trait map[string][]string) (err error) {
	typ := reflect.TypeOf(api)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		name := prefix + method.Name
		if _, ok := trait[method.Name]; !ok {
			continue
		}
		fn := method.Func.Interface()
		err = p.create(name, api, fn)
		if err != nil {
			return
		}
		p.fns[name] = reflect.ValueOf(fn)
		p.objs[name] = api
		p.trait[name] = trait[method.Name]
	}
	return
}

func (p *Server) create(name string, api interface{}, fn interface{}) (err error) {
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
	return
}

func (p *Server) Serve(w http.ResponseWriter, r *http.Request) {
	begin := time.Now().UnixNano()
	cost := begin
	cb := 0

	url := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	name := url[len(url) - 1]

	defer func() {
		end := time.Now().UnixNano()
		p.measure.Record("api.cost.handle." + name, cost - begin)
		p.measure.Record("api.cost.all." + name, end - begin)
		p.measure.Record("api.response." + name, int64(cb))
	}()

	var status string
	var detail string
	result, err := p.handle(name, w, r)
	cost = time.Now().UnixNano()
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

	cb = len(ret)

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

func (p *Server) handle(name string, w http.ResponseWriter, r *http.Request) (result []interface{}, err error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	if len(data) == 0 {
		return p.invoke(name, nil)
	}
	var in map[string]json.RawMessage
	err = json.Unmarshal(data, &in)
	if err != nil {
		return
	}

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
		if _, ok := args[argName]; !ok {
			return nil, NewErrRpcServer(fmt.Errorf("api %s arg %s missing", argName))
		} else {
			typ := fn.Type().In(i + 1)
			val := reflect.New(typ)
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

func NewServer() (p *Server) {
	p = &Server {
		make(map[string]reflect.Value),
		make(map[string]interface{}),
		make(map[string][]string),
		measure.NewMeasure(time.Second * 60, time.Second * 60 * 60 * 24),
	}

	err := p.reg("Measure.", p.measure, MeasureTrait)
	if err != nil {
		panic(err)
	}

	err = p.reg("Api.", p, BuiltinTrait)
	if err != nil {
		panic(err)
	}
	return
}

type Server struct {
	fns map[string]reflect.Value
	objs map[string]interface{}
	trait map[string][]string
	measure *measure.Measure
}

var MeasureTrait = map[string][]string {
	"Sync": {"time"},
}
var BuiltinTrait = map[string][]string {
	"List": {},
}

const (
	StatusOK = "OK"
	StatusErr = "error"
	HttpCodeOK = http.StatusOK
	HttpCodeErr = 599
)

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

type ErrRpcServer struct {
	err error
}

func NewErrRpcServer(e error) *ErrRpcServer {
	return &ErrRpcServer{e}
}

func (p *ErrRpcServer) Error() string {
	return "rpc server error: " + p.err.Error()
}

type ApiTrait interface {
	Trait() map[string][]string
}
