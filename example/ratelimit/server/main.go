/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

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
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// If you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
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
	svcName := "polaris.ratelimit.echo"
	Info := &registry.Info{
		ServiceName: svcName,
		Tags: map[string]string{
			polaris.NameSpaceTagKey: Namespace,
		},
	}

	qpsLimiter, err := polaris.NewQPSLimiter()
	if err != nil {
		log.Fatal(err)
	}
	qpsLimiter.WithNamespace(Namespace)
	qpsLimiter.WithServiceName(svcName)

	newServer := hello.NewServer(
		new(HelloImpl),
		server.WithRegistry(r),
		server.WithRegistryInfo(Info),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8890}),
		server.WithQPSLimiter(qpsLimiter),
	)

	err = newServer.Run()
	if err != nil {
		log.Fatal(err)
	}
}
