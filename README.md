文档：https://geektutu.com/post/geerpc-day1.html

> 定义2种Codec, `Gob`和`Json`, 但是实际代码中只实现了`Gob`一种, 事实上, 2者的实现非常接近, 甚至只需要把`gob`换成`json`即可。

## 通信过程

http报文

- header
- body
  - body的格式通过header 的`Content-Type` 和`Content-Length`
  - 解析header就能知道

RPC协议

- 这部分需要自主设计比如：第1个字节用来表示序列化方式，第2个字节表示压缩方式，第3-6个字节表示header的长度，7-10字节表示body的长度

服务端：消息编解码 放到结构体中`Option`中承载 `/server.go`

客户端：采用json编码Option，后续的header和body的编码方式由Option中的CodyType指定，服务端首先使用JSON解码Option，然后通过Option的CodeType解码剩余的内容。

```js
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
```

Option固定在报文的最开始，Header和Body可以有多个，即报文

```js
| Option | Header1 | Body1 | Header2 | Body2 | ...
```

`severCodec`  的过程非常简单，主要包含三个阶段

- 读取请求readRequest
- 处理请求handleRequest
- 回复请求sendRequest

> 这里用for无限制等待请求的到来，直到发生错误（例如连接被关闭， 接收到的报文有问题等）需要注意3点
>
> - handleRequest使用了协程并发执行请求
> - 处理请求是并发的，但是回复是必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析。这里使用锁（sending）保证。
> - 尽力而为，只有在header解析失败时，才终止循环

## day2

### call设计

- the method’s type is exported.
- the method is exported.
- the method has two arguments, both exported (or builtin) types.
- the method’s second argument is a pointer.
- the method has return type error.

例如：

````js
func (t *T) MethodName(argType T1, replyType *T2) error
````

## day3

#### 集成到服务端

- 通过反射结构体已经映射为服务，但请求过程还没有完成。
- 从接收到请求到回复还差以下几步
  - 第一步，根据入参类型，将请求的body反序列化；
  - 第二步，调用`service.call` ，完成方法调用；
  - 第三步，将reply序列化为字节流，构造为字节流，构造响应报文，返回。

## day4（超时处理）

超时处理是RPC框架中一个比较基本能力，

需要客户端处理超时的地方有：

- 与服务端建立连接，导致的超时
- 发送请求到服务端，写报文导致的超时
- 等待服务端处理时，等待处理导致的超时，（比如服务端已挂死，迟迟不响应）
- 从服务端接受响应时，读报文导致的超时

需要服务端处理超时的地方有：

- 读取客户端请求报文时，读报文导致的超时
- 发送响应报文时，写报文导致的超时
- 调用映射服务的方法时，处理报文导致的超时
-
