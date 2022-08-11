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
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/polaris"
)

func main() {
	newClient := hello.MustNewClient("polaris.quickstart.echo",
		client.WithSuite(polaris.NewDefaultClientSuite()), // you can also refer readme to customize the initialization of each component
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
