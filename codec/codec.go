package codec

import "io"

type Header struct {
	ServiceMethod string //format "Service.Method"
	Seq          uint64 //sequence number chosen by client
	Error        string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecMap map[Type]NewCodecFunc

// 初始化
func init() {
	NewCodecMap = make(map[Type]NewCodecFunc)
	NewCodecMap[GobType] = NewGobCodec
}
