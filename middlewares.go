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
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/polarismesh/polaris-go/pkg/model"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/polarismesh/polaris-go/api"
)

func NewUpdateServiceCallResultMW(configFile ...string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request, response interface{}) error {
			retCode := int32(RetSuccessCode)
			retStatus := api.RetSuccess
			begin := time.Now()
			kitexCallErr := next(ctx, request, response)
			cost := time.Since(begin)
			if kitexCallErr != nil {
				retCode = RetFailCode
				retStatus = api.RetFail
			}

			ri := rpcinfo.GetRPCInfo(ctx)
			sdkCtx, err := GetPolarisConfig(configFile...)
			if err != nil {
				return err
			}
			consumer := api.NewConsumerAPIByContext(sdkCtx)
			ns, _ := ri.To().Tag(NameSpaceKey)
			instanceId, ok := ri.To().Tag(InstanceIDKey)
			if !ok {
				// 没有找到实例
				return kitexCallErr
			}

			req := api.InstanceRequest{
				ServiceKey: model.ServiceKey{
					Namespace: ns,
					Service:   ri.To().ServiceName(),
				},
				InstanceID: instanceId,
			}

			svcCallResult, reportErr := api.NewServiceCallResult(sdkCtx, req)
			if reportErr != nil {
				return reportErr
			}

			svcCallResult.SetRetCode(retCode)
			svcCallResult.SetRetStatus(retStatus)
			svcCallResult.SetDelay(cost)
			// 执行调用结果上报
			_ = consumer.UpdateServiceCallResult(svcCallResult)
			return kitexCallErr
		}
	}
}
