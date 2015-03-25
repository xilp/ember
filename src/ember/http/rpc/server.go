package rpc

import (
	"reflect"
	"sync"
	"net/http"
	"io/ioutil"
)

const (
	StatusErr = 599
	ErrOK = "OK"
) 

type Server struct {
	mu sync.Mutex
	
	addr string
	funcs map[string]reflect.Value

	running  bool
	objs interface{}
}

func NewServer(addr string) *Server {
	RegisterType(RpcError{})

	s := new(Server)
	s.mu = &sync.Mutex{} 

	s.network = network
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
		err = s.Register(mname, method.Func.Interface())		
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

	mu.Lock()
	if _, ok := s.funcs[name]; ok {
		err = fmt.Errorf("%s has registered", name)
		mu.Unlock()
		return
	}

	s.funcs[name] = v
	mu.Unlock()
	return
}

func (s *Server) start() error {
	http.HandleFunc("/Router", routeTo)
	return http.ListenAndServe("localhost", nil)
}

func (s *Server) routeTo(w http.ResponseWriter, r *http.Request) {
	var errStr string
	result, err := doRoute(w, r)
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
		w.WriteHeader(StatusErr)
	} else {
		w.WriteHeader(http.StatusOK)
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

	if dataMap, ok := f.(map[string]interface{}); !ok {
		err = fmt.Errorf("data structure type not match")
		return
	}

	name := r.URL.Path
	if name == nil {
		err = fmt.Errorf("rpc name is null")	
		return 
	}

	array, ok := dataMap["args"]
	if args, ok := array.([]interface{}); !ok {
		err = fmt.Errorf("data structure type not match")
		return
	}
	
	return s.handle(name, args);
}

func (s *Server) handle(name string, args []interface{}) ([]interface{}, error) {
	mu.Lock()
	f, ok := s.funcs[name]
	mu.Unlock()
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

func NewResponse(err string, result []byte) *Response {
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

