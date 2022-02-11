package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
)

type serverFunc func(interface{}) interface{}
type rpcHandle struct {
	handler  serverFunc
	argsType reflect.Type
	resType  reflect.Type
}
type RPCServer struct {
	router map[string]rpcHandle
}

//rpc客户端请求的数据
type request struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

func Server() RPCServer {
	return RPCServer{make(map[string]rpcHandle)}
}

//Register 注册服务端方法，服务端需实现两个函数，其中handler用于获取句柄，service用于获取实际参数类型.
func (r *RPCServer) Register(name string, handler serverFunc, service interface{}) error {
	return r.register(name, handler, service)
}

func (r *RPCServer) ListenAndServe(address string) error {
	listener, err := r.listen(address)
	if err != nil {
		return err
	}
	if err = r.accept(listener); err != nil {
		return err
	}
	return nil
}

func (r *RPCServer) register(name string, handler serverFunc, service interface{}) error {
	serviceType := reflect.TypeOf(service)
	if err := r.checkHandlerType(serviceType); err != nil {
		return err
	}
	argsType := serviceType.In(0)
	resType := serviceType.Out(0)
	r.router[name] = rpcHandle{handler: handler, argsType: argsType, resType: resType}
	return nil
}

func (r *RPCServer) checkHandlerType(handlerType reflect.Type) error {
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler is not func")
	}
	if handlerType.NumIn() != 1 || handlerType.NumOut() != 1 {
		return fmt.Errorf("handler input/output parameters number must be 1")
	}
	if handlerType.In(0).Kind() != reflect.Struct || handlerType.Out(0).Kind() != reflect.Struct {
		return fmt.Errorf("handler input/output parameters must be Struct")
	}
	return nil
}

func (r *RPCServer) listen(address string) (*net.TCPListener, error) {
	ladder, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	listener, err := net.ListenTCP("tcp4", ladder)
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	return listener, nil
}

func (r *RPCServer) accept(listener *net.TCPListener) error {
	defer listener.Close()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			return err
		}
		defer conn.Close()
		go r.handle(conn)
	}
}

func (r *RPCServer) handle(conn *net.TCPConn) {
	strLen := make([]byte, PackMaxSize)
	for {
		n, err := conn.Read(strLen)
		if err != nil && err != io.EOF {
			log.Println("read data len failed. err:" + err.Error())
		}
		if n <= 0 {
			log.Println("no data")
		}
		len, _ := strconv.ParseInt(string(strLen[:PackMaxSize]), 10, 64)

		// buff存放实际data部分
		buff := make([]byte, len)
		n, err = conn.Read(buff)
		if err != nil {
			log.Println("read data info failed. err:" + err.Error())
			return
		}
		if n <= 0 {
			log.Println("no data")
			return
		}

		//调度,处理实际的内容.
		rsp, err := r.dispatcher(buff)
		if err != nil {
			log.Println("dispatch failed. err:" + err.Error())
		}
		rspBytes, err := r.packResponse(rsp)
		if err != nil {

		}
		conn.Write(rspBytes)
	}
}

func (r *RPCServer) dispatcher(req []byte) (interface{}, error) {
	var cReq request
	if err := json.Unmarshal(req, &cReq); err != nil {
		return nil, err
	}
	rh, ok := r.router[cReq.Name]
	if !ok {
		return nil, fmt.Errorf("cannot find handler named %s", cReq.Name)
	}

	args := reflect.New(rh.argsType).Interface()
	if err := json.Unmarshal(cReq.Data, args); err != nil {
		return nil, err
	}
	return rh.handler(args), nil
}

func (r *RPCServer) packResponse(v interface{}) ([]byte, error) {
	return pack(v)
}
