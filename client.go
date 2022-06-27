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

package polaris

import (
	"log"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/loadbalance"
)

type ClientOptions struct {
	DstMetadata  map[string]string `json:"dst_metadata"`
	SrcNamespace string            `json:"src_namespace"`
	SrcService   string            `json:"src_service"`
	SrcMetadata  map[string]string `json:"src_metadata"`
}

// ClientSuite client的一些配置信息
type ClientSuite struct {
	DstNameSpace       string                   // 目标服务空间,用于服务发现
	Resolver           discovery.Resolver       // 服务发现
	Balancer           loadbalance.Loadbalancer // 负载均衡
	ReportCallResultMW endpoint.Middleware      // 服务结果上报中间件,用于熔断
}

// Options implements the client.Suite interface.
func (cs *ClientSuite) Options() []client.Option {
	var (
		// options
		opts []client.Option
	)

	if len(cs.DstNameSpace) < 0 {
		cs.DstNameSpace = DefaultPolarisNamespace
	}
	opts = append(opts, client.WithTag(DstNameSpaceTagKey, cs.DstNameSpace))

	if cs.Resolver == nil {
		o := ClientOptions{}
		r, err := NewPolarisResolver(o)
		if err != nil {
			log.Fatal(err)
		}
		cs.Resolver = r
	}
	opts = append(opts, client.WithResolver(cs.Resolver))

	if cs.Balancer != nil {
		opts = append(opts, client.WithLoadBalancer(cs.Balancer))
	}

	if cs.ReportCallResultMW != nil {
		opts = append(opts, client.WithMiddleware(cs.ReportCallResultMW))
	}

	return opts
}
