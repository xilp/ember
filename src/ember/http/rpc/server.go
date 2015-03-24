package rpc

import (
	"reflect"
	"sync"
	"net/http"
	"io/ioutil"
)

type Server struct {
	sync.Mutex
	
	network string
	addr string
	funcs map[string]reflect.Value

	listener net.Listener
	running  bool
	objs interface{}
}

func NewServer(network, addr string) *Server {
	RegisterType(RpcError{})

	s := new(Server)
	s.network = network
	s.addr = addr

	s.funcs = make(map[string]reflect.Value)
	return s
}

func (s *Server) Run() error {
	var err error
	s.listener, err = net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}

	s.running = true

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.onConn(conn)
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

func routeTo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println("read body error", err.Error())
	}

	config := &Config{}
	err := json.Unmarshal(data, config)
	if err != nil {
		println("unmarshal err", err.Error())
	}

	name := config.Name
	args := config.Args
	

		
	
}

func (s *Server) handle(name string, args []interface{}) ([]byte, error) {
	name, args, err := decodeData(data)
	if err != nil {
		return nil, err
	}

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

	return encodeData(name, outArgs)
}

type Config struct {
	Name string `json:"name"`
	Args []interface{} `json:"args"`
}







