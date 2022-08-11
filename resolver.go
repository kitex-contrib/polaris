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
	"fmt"
	"strings"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/log"
	"github.com/polarismesh/polaris-go/pkg/model"
)

const (
	DefaultWeight = 10
)

// Resolver is extension interface of Kitex discovery.Resolver.
type Resolver interface {
	discovery.Resolver
	Watcher(ctx context.Context, desc string) (discovery.Change, error)
}

// polarisResolver is a resolver using polaris.
type polarisResolver struct {
	provider api.ProviderAPI
	consumer api.ConsumerAPI
	o        ClientOptions
}

// NewPolarisResolver creates a polaris based resolver.
func NewPolarisResolver(o ClientOptions, configFile ...string) (Resolver, error) {
	sdkCtx, err := GetPolarisConfig(configFile...)
	if err != nil {
		return nil, err
	}

	newInstance := &polarisResolver{
		consumer: api.NewConsumerAPIByContext(sdkCtx),
		provider: api.NewProviderAPIByContext(sdkCtx),
		o:        o,
	}

	return newInstance, nil
}

// Target implements the Resolver interface.
func (pr *polarisResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	// serviceName identification is generated by namespace and serviceName to identify serviceName
	var serviceIdentification strings.Builder

	dstNamespace, ok := target.Tag(DstNameSpaceTagKey)
	if ok {
		serviceIdentification.WriteString(dstNamespace)
	} else {
		serviceIdentification.WriteString(DefaultPolarisNamespace)
	}
	serviceIdentification.WriteString(":")
	serviceIdentification.WriteString(target.ServiceName())

	return serviceIdentification.String()
}

// Watcher return registered service changes.
func (pr *polarisResolver) Watcher(ctx context.Context, desc string) (discovery.Change, error) {
	var (
		eps    []discovery.Instance
		add    []discovery.Instance
		update []discovery.Instance
		remove []discovery.Instance
	)
	namespace, serviceName := SplitDescription(desc)
	key := model.ServiceKey{
		Namespace: namespace,
		Service:   serviceName,
	}
	watchReq := api.WatchServiceRequest{}
	watchReq.Key = key
	watchRsp, err := pr.consumer.WatchService(&watchReq)
	if nil != err {
		log.GetBaseLogger().Fatalf("fail to WatchService, err is %v", err)
	}
	instances := watchRsp.GetAllInstancesResp.Instances

	if nil != instances {
		for _, instance := range instances {
			log.GetBaseLogger().Infof("instance getOneInstance is %s:%d", instance.GetHost(), instance.GetPort())
			eps = append(eps, ChangePolarisInstanceToKitex(instance, pr.o))
		}
	}

	result := discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}
	Change := discovery.Change{}

	select {
	case <-ctx.Done():
		log.GetBaseLogger().Infof("[Polaris resolver] Watch has been finished")
		return Change, nil
	case event := <-watchRsp.EventChannel:
		eType := event.GetSubScribeEventType()
		if eType == api.EventInstance {
			insEvent := event.(*model.InstanceEvent)
			if insEvent.AddEvent != nil {
				for _, instance := range insEvent.AddEvent.Instances {
					add = append(add, ChangePolarisInstanceToKitex(instance, pr.o))
				}
			}
			if insEvent.UpdateEvent != nil {
				for i := range insEvent.UpdateEvent.UpdateList {
					update = append(update, ChangePolarisInstanceToKitex(insEvent.UpdateEvent.UpdateList[i].After, pr.o))
				}
			}
			if insEvent.DeleteEvent != nil {
				for _, instance := range insEvent.DeleteEvent.Instances {
					remove = append(remove, ChangePolarisInstanceToKitex(instance, pr.o))
				}
			}
			Change = discovery.Change{
				Result:  result,
				Added:   add,
				Updated: update,
				Removed: remove,
			}
		}
		return Change, nil
	}
}

// Resolve implements the Resolver interface.
func (pr *polarisResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	var eps []discovery.Instance
	namespace, serviceName := SplitDescription(desc)
	getInstances := &api.GetInstancesRequest{}
	getInstances.Namespace = namespace
	getInstances.Service = serviceName
	InstanceResp, err := pr.consumer.GetInstances(getInstances)
	if nil != err {
		log.GetBaseLogger().Fatalf("fail to getOneInstance, err is %v", err)
	}
	instances := InstanceResp.GetInstances()
	if nil != instances {
		for _, instance := range instances {
			log.GetBaseLogger().Infof("instance getOneInstance is %s:%d", instance.GetHost(), instance.GetPort())
			eps = append(eps, ChangePolarisInstanceToKitex(instance, pr.o))
		}
	}

	if len(eps) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %s", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

// Diff implements the Resolver interface.
func (pr *polarisResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name implements the Resolver interface.
func (pr *polarisResolver) Name() string {
	return "Polaris"
}
