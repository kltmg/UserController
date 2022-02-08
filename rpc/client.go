package rpc

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"
)

//connection pool
type RPCClient struct {
	pool chan net.TCPConn
}

func Client(connections int, address string) (RPCClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return RPCClient{}, err
	}
	pool := make(chan net.TCPConn, connections)
	for i := 0; i < connections; i++ {
		conn, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			return RPCClient{}, err
		}
		pool <- *conn
	}
	return RPCClient{pool: pool}, nil
}

//Call 对外提供方法， 调用服务端方法， resp必须为指针类型，保存返回结果数据.
func (r *RPCClient) Call(name string, req interface{}, resp interface{}) error {
	return r.call(name, req, resp)
}

//call 真正rpc调用逻辑，  使用rpc调用函数name(req), 并将结果保存到resp中.
func (r *RPCClient) call(name string, req interface{}, resp interface{}) error {
	conn := r.getConn()
	defer r.releaseConn(conn)

	reqBytes, err := r.packRequest(name, req)
	if err != nil {
		return err
	}

	conn.Write(reqBytes)

	strLen := make([]byte, PackMaxSize)
	n, err := conn.Read(strLen)
	if err != nil && err != io.EOF {
		log.Panicln("read data len failed. err:" + err.Error())
		return err
	}
	if n <= 0 {
		log.Panicln("no data")
		return err
	}
	len, _ := strconv.ParseInt(string(strLen[:PackMaxSize]), 10, 64)

	// buff存放实际data部分
	buff := make([]byte, len)
	n, err = conn.Read(buff)
	if err != nil {
		log.Panicln("read data info failed. err:" + err.Error())
		return err
	}
	if n <= 0 {
		log.Panicln("no data")
	}
	if err := json.Unmarshal(buff, resp); err != nil {
		return err
	}
	return nil
}

func (r *RPCClient) getConn() net.TCPConn {
	select {
	case conn := <-r.pool:
		return conn
	}
}

func (r *RPCClient) releaseConn(conn net.TCPConn) {
	select {
	case r.pool <- conn:
		return
	}
}

func (r *RPCClient) packRequest(name string, v interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	reqSt := request{Name: name, Data: dataBytes}
	reqBytes, err := pack(reqSt)
	if err != nil {
		return nil, err
	}
	return reqBytes, nil
}
