package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"ember/measure"
)

func (p *Server) Run(path string, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc(path, p.Serve)
	return http.ListenAndServe(":" + strconv.Itoa(port), mux)
}

func (p *Server) Serve(w http.ResponseWriter, r *http.Request) {
	begin := time.Now().UnixNano()
	cost := begin
	cb := 0

	url := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	name := url[len(url) - 1]

	var err error
	defer func() {
		end := time.Now().UnixNano()
		p.measure.Record("api.cost.handle." + name, (cost - begin) / 1000)
		p.measure.Record("api.cost.all." + name, (end - begin) / 1000)
		p.measure.Record("api.resp." + name, int64(cb))
		if err != nil {
			p.measure.Record("api.error." + name, 0)
		}
	}()

	var data []byte
	data, err = p.handle(name, w, r)
	cost = time.Now().UnixNano()
	cb = len(data)

	h := w.Header()
	h.Set("Content-Type", "text/json")

	if err == nil {
		w.WriteHeader(HttpCodeOK)
	} else {
		w.WriteHeader(HttpCodeErr)
	}

	_, err = w.Write(data)
}

func (p *Server) handle(name string, w http.ResponseWriter, r *http.Request) (data []byte, err error) {
	var status string
	var detail string

	result, err := p.call(name, w, r)
	if err == nil {
		status = StatusOK
	} else {
		status = StatusErr
		result = nil
		detail = NewErrRpcFailed(err).Error()
	}

	resp := NewResponse(status, detail, result)
	data, err = json.Marshal(resp)
	if err != nil {
		resp = NewResponse(StatusErr, NewErrRpcFailed(err).Error(), nil)
		data, err = json.Marshal(resp)
	}
	return
}

func (p *Server) call(name string, w http.ResponseWriter, r *http.Request) (ret []interface{}, err error) {
	fn, ok1 := p.fns[name]
	fv, ok2 := p.fvs[name]
	if !ok1 || !ok2 {
		err = fmt.Errorf("%s not found", name)
		return
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	ret, err = fv.Invoke(fn.proto, data)
	return
}

func (p *Server) List() (protos []FnProto, err error) {
	protos = p.fns.List()
	return
}

func (p *Server) Reg(sobj interface{}, cobj interface{}) (err error) {
	return p.reg("", sobj, cobj)
}

func (p *Server) reg(prefix string, sobj interface{}, cobj interface{}) (err error) {
	fns := NewFnTraits(cobj)
	fvs, err := NewFnValues(sobj, fns)
	if err != nil {
		return
	}
	for origin, fn := range fns {
		name := prefix + origin
		if _, ok := p.fns[name]; ok {
			err = fmt.Errorf("%s has registered", name)
			return
		}
		fn.Prefix(prefix)
		p.fns[name] = fn
		if _, ok := fvs[origin]; !ok {
			continue
		}
		p.fvs[name] = fvs[origin]
	}
	return
}

func NewServer() (p *Server) {
	p = &Server {
		make(FnTraits),
		make(FnValues),
		measure.NewMeasure(time.Second * 60, time.Second * 60 * 60 * 24),
	}

	err := p.reg(MeasurePrefix, p.measure, &Measure{})
	if err != nil {
		panic(err)
	}

	err = p.reg(BuiltinPrefix, p, &Builtin{})
	if err != nil {
		panic(err)
	}
	return
}

type Server struct {
	fns FnTraits
	fvs FnValues
	measure *measure.Measure
}

func (p *ErrRpcFailed) Error() string {
	return "[rpc server] " + p.err.Error()
}

func NewErrRpcFailed(err error) error {
	if _, ok := err.(*ErrCallFailed); ok {
		return err
	}
	return &ErrRpcFailed{err}
}

type ErrRpcFailed struct {
	err error
}

func NewResponse(status, detail string, result []interface{}) *Response {
	return &Response{status, detail, result}
}

type Response struct {
	Status string        "status"
	Detail string        "detail"
	Result []interface{} "result"
}

const (
	StatusOK = "OK"
	StatusErr = "error"
	HttpCodeOK = http.StatusOK
	HttpCodeErr = 599
)
