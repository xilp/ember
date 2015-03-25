package rpc

import (
	"reflect"
	"sync"
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

const (
	StatusErr = 599
	ErrOK = "OK"
) 

type Server struct {
	sync.Mutex
	
	addr string
	funcs map[string]reflect.Value

	running  bool
	obj interface{}
}

func NewServer(addr string) *Server {
	s := new(Server)
	s.addr = addr

	s.funcs = make(map[string]reflect.Value)
	
	return s
}

func (s *Server) Run() error {
	err := s.start()
	if err != nil {
		return err
	}
	
	return nil
}

func (s *Server) RegisterObj(obj interface{}) (err error) {
	typ := reflect.TypeOf(obj)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		mname := method.Name
		err = s.register(mname, method.Func.Interface())		
		if err != nil {
			return
		}
	}

	s.obj = obj
	return	
}

func (s *Server) register(name string, f interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s is not callable", name)
		}
	}()

	v := reflect.ValueOf(f)

	//to check f is function
	v.Type().NumIn()

	nOut := v.Type().NumOut()
	if nOut == 0 || v.Type().Out(nOut-1).Kind() != reflect.Interface {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	_, b := v.Type().Out(nOut - 1).MethodByName("Error")
	if !b {
		err = fmt.Errorf("%s return final output param must be error interface", name)
		return
	}

	s.Lock()
	if _, ok := s.funcs[name]; ok {
		err = fmt.Errorf("%s has registered", name)
		s.Unlock()
		return
	}

	s.funcs[name] = v
	s.Unlock()
	return
}

func (s *Server) start() error {
	println("Server.Start")
	http.HandleFunc("/Router", s.routeTo)
	return http.ListenAndServe(":11182", nil)
}

func (s *Server) routeTo(w http.ResponseWriter, r *http.Request) {
	println("hahahahahahahahha")
	var errStr string
	result, err := s.doRoute(w, r)
	if err == nil {
		errStr = ErrOK
	} else {
		errStr = err.Error()
	}
		
	resp := NewResponse(errStr, result)
	ret, mErr := json.Marshal(resp)
	if mErr != nil {
		errStr = "marshal error"
	}
	
	h := w.Header()
	h.Set("Content-Type", "text/json")

	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(StatusErr)
	}
	_, err = w.Write(ret)
}

func (s *Server) doRoute(w http.ResponseWriter, r *http.Request) (result []interface{}, err error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	
	var f interface{}
	err = json.Unmarshal(data, &f)
	if err != nil {
		return
	}

	dataMap, ok := f.(map[string]interface{});
	if !ok {
		err = fmt.Errorf("data structure type not match")
		return
	}

	//name := r.URL.Path
	name := "Larger"
	array, ok := dataMap["args"]
	args, ok := array.([]interface{});
	if !ok {
		err = fmt.Errorf("data structure type not match")
		return
	}
	
	return s.handle(name, args);
}

func (s *Server) handle(name string, args []interface{}) ([]interface{}, error) {
	s.Lock()
	f, ok := s.funcs[name]
	s.Unlock()
	if !ok {
		return nil, fmt.Errorf("rpc %s not registered", name)
	}
	
	inValues := make([]reflect.Value, len(args) + 1)
	inValues[0] = reflect.ValueOf(s.obj)

	for i := 0; i < len(args); i++ {
		
		if args[i] == nil {
			inValues[i + 1] = reflect.Zero(f.Type().In(i))
		} else {
			
			inValues[i + 1] = reflect.ValueOf(args[i])
		}
	}

	out := f.Call(inValues)
	
	outArgs := make([]interface{}, len(out))
	for i := 0; i < len(outArgs); i++ {
		outArgs[i] = out[i].Interface()
	}

	p := out[len(out)-1].Interface()
	if p != nil {
		if e, ok := p.(error); ok {
			outArgs[len(out)-1] = RpcError{e.Error()}
		} else {
			return nil, fmt.Errorf("final param must be error")
		}
	}

	return outArgs, nil
}

func NewData(name string, args []interface{}) *Data {
	d := new(Data)
	d.Name = name
	d.Args = args
	return d 	
}

func NewResponse(err string, result []interface{}) *Response {
	r := new(Response)
	r.Err = err
	r.Result = result
	return r
}	

type Data struct {
	Name string `json:"name"`
	Args []interface{} `json:"args"`
}

type Response struct {
	Err string	`json:"err"`
	Result []interface{}	`json:"result"`
}

type RpcError struct {
	Message string
}

func (r RpcError) Error() string {
	return r.Message
}
