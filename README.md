文档：https://geektutu.com/post/geerpc-day1.html

> 定义 2 种 Codec, `Gob`和 `Json`, 但是实际代码中只实现了 `Gob`一种, 事实上, 2 者的实现非常接近, 甚至只需要把 `gob`换成 `json`即可。

## 通信过程

http 报文

-   header
-   body
    -   body 的格式通过 header 的 `Content-Type` 和 `Content-Length`
    -   解析 header 就能知道

RPC 协议

-   这部分需要自主设计比如：第 1 个字节用来表示序列化方式，第 2 个字节表示压缩方式，第 3-6 个字节表示 header 的长度，7-10 字节表示 body 的长度

服务端：消息编解码 放到结构体中 `Option`中承载 `/server.go`

客户端：采用 json 编码 Option，后续的 header 和 body 的编码方式由 Option 中的 CodyType 指定，服务端首先使用 JSON 解码 Option，然后通过 Option 的 CodeType 解码剩余的内容。

```js
| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
```

Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文

```js
| Option | Header1 | Body1 | Header2 | Body2 | ...
```

`severCodec` 的过程非常简单，主要包含三个阶段

-   读取请求 readRequest
-   处理请求 handleRequest
-   回复请求 sendRequest

> 这里用 for 无限制等待请求的到来，直到发生错误（例如连接被关闭， 接收到的报文有问题等）需要注意 3 点
>
> -   handleRequest 使用了协程并发执行请求
> -   处理请求是并发的，但是回复是必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析。这里使用锁（sending）保证。
> -   尽力而为，只有在 header 解析失败时，才终止循环

## day2

### call 设计

-   the method’s type is exported.
-   the method is exported.
-   the method has two arguments, both exported (or builtin) types.
-   the method’s second argument is a pointer.
-   the method has return type error.

例如：

```js
func (t *T) MethodName(argType T1, replyType *T2) error
```

## day3

#### 集成到服务端

-   通过反射结构体已经映射为服务，但请求过程还没有完成。
-   从接收到请求到回复还差以下几步
    -   第一步，根据入参类型，将请求的 body 反序列化；
    -   第二步，调用 `service.call` ，完成方法调用；
    -   第三步，将 reply 序列化为字节流，构造为字节流，构造响应报文，返回。

## day4（超时处理）

超时处理是 RPC 框架中一个比较基本能力，

需要客户端处理超时的地方有：

-   与服务端建立连接，导致的超时
-   发送请求到服务端，写报文导致的超时
-   等待服务端处理时，等待处理导致的超时，（比如服务端已挂死，迟迟不响应）
-   从服务端接受响应时，读报文导致的超时

需要服务端处理超时的地方有：

-   读取客户端请求报文时，读报文导致的超时
-   发送响应报文时，写报文导致的超时
-   调用映射服务的方法时，处理报文导致的超时

## day5 支持 HTTP 协议

通信过程

1。客户端向 RPC 服务器发送 CONNECT 请求

`CONNECT 10.0.0.1:9999/_Grpc_ HTTP/1.0`

2。RPC 服务器返回 HTTP 200 状态码表示连接建立。

`HTTP/1.0 200 Connected to Grpc `

## day6 负载均衡策略

> XClient 的构造函数需要传入三个参数，服务发现实例 `Discovery`负载均衡模式 `SelectMode`以及协议选项 `Option`。为了尽量地复用已经创建好的 Socket 连接。连接 clients 保存创建成功的 Client 实例，并提供 Close 方法在结束后，关闭已经建立的连接
