package Grpc

import (
	"Grpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Call represents an active RPC
type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} //aruments to the function
	Reply         interface{} //reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete
}

func (call *Call) done() {
	call.Done <- call
}

/*
*

	Client represents an RPC Client
	There may be multiple outstanding Calls associated
	with a single Client, and a Client may be used be
	mulitple goroutines simultaneously
*/
type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header // protect following
	mu       sync.Mutex   // protect following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // sever has told us to stop

}

var _ io.Closer = (*Client)(nil)
var ErrShutdown = errors.New("connection is shut down")

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

/*
*

	registerCall 将参数call 添加到client.pending 中，并更新client.seq
	removeCall 根据seq， 从client.pending中移出对应的call，并返回
	terminateCalls 服务端或客户端发送错误时调用，将shutdown设置为true，且将错误信息通知所有pending状态的call
*/
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}

	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// if usually means that Write partially failed
			// and call was already removed
			err = client.cc.ReadBody(nil)

		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()

		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// error occurs, so terminateCalls pending calls
	client.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	// send options with server
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error:", err)
		_ = conn.Close()
		return nil, err
	}
	return NewClientCodec(f(conn), opt), nil
}

func NewClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}
