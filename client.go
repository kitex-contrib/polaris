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

// ClientSuite It is used to assemble multiple associated client's Options
type ClientSuite struct {
	DstNameSpace       string                   // dest namespace for service discovery
	Resolver           discovery.Resolver       // service discovery component
	Balancer           loadbalance.Loadbalancer // load balancer
	ReportCallResultMW endpoint.Middleware      // report service call result for circuitbreak
}

func NewDefaultClientSuite() *ClientSuite {
	return &ClientSuite{}
}

// Options implements the client.Suite interface.
func (cs *ClientSuite) Options() []client.Option {
	var opts []client.Option

	if len(cs.DstNameSpace) <= 0 {
		cs.DstNameSpace = DefaultPolarisNamespace
	}
	opts = append(opts, client.WithTag(DstNameSpaceTagKey, cs.DstNameSpace))

	var resolver discovery.Resolver
	if cs.Resolver != nil {
		resolver = cs.Resolver
	} else {
		o := ClientOptions{}
		r, err := NewPolarisResolver(o)
		if err != nil {
			log.Fatal(err)
		}
		resolver = r
	}
	opts = append(opts, client.WithResolver(resolver))

	var lb loadbalance.Loadbalancer
	if cs.Balancer != nil {
		lb = cs.Balancer
	} else {
		pb, err := NewPolarisBalancer()
		if err != nil {
			log.Fatal(err)
		}
		lb = pb
	}
	opts = append(opts, client.WithLoadBalancer(lb))

	if cs.ReportCallResultMW != nil {
		opts = append(opts, client.WithMiddleware(cs.ReportCallResultMW))
	} else {
		opts = append(opts, client.WithMiddleware(NewUpdateServiceCallResultMW()))
	}

	return opts
}
