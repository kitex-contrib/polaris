# polaris（这是一个社区驱动的项目）

[English](https://github.com/kitex-contrib/polaris/blob/main/README.md)

[polaris](https://github.com/polarismesh/polaris) for [Kitex](https://github.com/cloudwego/kitex)

# 特性

## 服务注册与发现
- [x] 支持服务注册与发现

## 服务熔断
- [x] 支持服务熔断

## 动态路由
- [x] 支持动态路由

## 服务限流
- [x] 支持限流

# 服务端使用示例
```go
import (
	"context"
	"log"
	"net"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/polaris"
)

const (
	Namespace = "Polaris"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message + "Hi,Kitex!",
	}
	return resp, nil
}

func main() {
	so := polaris.ServerOptions{}
	r, err := polaris.NewPolarisRegistry(so)
	if err != nil {
		log.Fatal(err)
	}
	Info := &registry.Info{
		ServiceName: "polaris.quickstart.echo",
		Tags: map[string]string{
			polaris.NameSpaceTagKey: Namespace,
		},
	}
	newServer := hello.NewServer(
		new(HelloImpl),
		server.WithRegistry(r),
		server.WithRegistryInfo(Info),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8890}),
	)

	err = newServer.Run()
	if err != nil {
		log.Fatal(err)
	}
}
```

# 客户端使用示例
- 提供了2种方式，可以通过suite快速开始，也可以自定义初始化各个组件开始

## suite方式
```go
import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/polaris"
)

func main() {
	newClient := hello.MustNewClient("polaris.quickstart.echo",
		client.WithSuite(polaris.NewDefaultClientSuite()),
		client.WithRPCTimeout(time.Second*360),
	)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*360)
		resp, err := newClient.Echo(ctx, &api.Request{Message: "Hi,polaris!"})
		cancel()
		if err != nil {
			log.Println(err)
		}
		log.Println(resp)
		time.Sleep(1 * time.Second)
	}
}
```

## 自定义初始化各个组件方式
```go
import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/polaris"
)

const (
	Namespace = "Polaris"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// if you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	o := polaris.ClientOptions{}
	r, err := polaris.NewPolarisResolver(o)
	if err != nil {
		log.Fatal(err)
	}

	pb, err := polaris.NewPolarisBalancer()
	if err != nil {
		log.Fatal(err)
	}

	suite := &polaris.ClientSuite{
		DstNameSpace:       Namespace,
		Resolver:           r,
		Balancer:           pb,
		ReportCallResultMW: polaris.NewUpdateServiceCallResultMW(),
	}

	newClient := hello.MustNewClient("polaris.quickstart.echo",
		client.WithSuite(suite),
		client.WithRPCTimeout(time.Second*360),
	)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*360)
		resp, err := newClient.Echo(ctx, &api.Request{Message: "Hi,polaris!"})
		cancel()
		if err != nil {
			log.Println(err)
		}
		log.Println(resp)
		time.Sleep(1 * time.Second)
	}
}
```

# 更多信息

参考example
