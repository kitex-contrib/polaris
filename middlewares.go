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
	"context"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpcinfo/remoteinfo"
	"github.com/polarismesh/polaris-go/api"
)

const (
	retSuccessCode = 0
	retFailCode    = -1
)

// NewUpdateServiceCallResultMW report call result for circuitbreak
func NewUpdateServiceCallResultMW(configFile ...string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		var (
			callResultSdkCtx      api.SDKContext
			callResultConsumerAPI api.ConsumerAPI
			callResultOnce        sync.Once
		)
		var err error
		callResultOnce.Do(func() {
			callResultSdkCtx, err = GetPolarisConfig(configFile...)
			callResultConsumerAPI = api.NewConsumerAPIByContext(callResultSdkCtx)
		})
		if err != nil {
			return func(ctx context.Context, req, resp interface{}) (err error) {
				return err
			}
		}
		return func(ctx context.Context, request, response interface{}) error {
			retCode := int32(retSuccessCode)
			retStatus := api.RetSuccess
			begin := time.Now()
			kitexCallErr := next(ctx, request, response)
			cost := time.Since(begin)
			if kitexCallErr != nil {
				retCode = retFailCode
				retStatus = api.RetFail
			}

			svcCallResult := &api.ServiceCallResult{}

			ri := rpcinfo.GetRPCInfo(ctx)
			svcCallResult.CalledInstance = ri.To().(remoteinfo.RemoteInfo).GetInstance().(*polarisKitexInstance).polarisInstance

			svcCallResult.SetRetCode(retCode)
			svcCallResult.SetRetStatus(retStatus)
			svcCallResult.SetDelay(cost)
			// 执行调用结果上报
			_ = callResultConsumerAPI.UpdateServiceCallResult(svcCallResult)
			return kitexCallErr
		}
	}
}
