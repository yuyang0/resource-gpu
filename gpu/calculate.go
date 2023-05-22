package gpu

import (
	"context"
	"strings"

	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	gputypes "github.com/projecteru2/resource-gpu/gpu/types"
)

// CalculateDeploy .
func (p Plugin) CalculateDeploy(ctx context.Context, nodename string, deployCount int, resourceRequest plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.gpu.CalculateDeploy").WithField("node", nodename)
	req := &gputypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.WithField("node", nodename).Error(ctx, err)
		return nil, err
	}

	var enginesParams []*gputypes.EngineParams
	var workloadsResource []*gputypes.WorkloadResource

	enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, deployCount, req)
	if err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"engines_params":     enginesParams,
		"workloads_resource": workloadsResource,
	}, nil
}

// CalculateRealloc .
func (p Plugin) CalculateRealloc(ctx context.Context, nodename string, resource plugintypes.WorkloadResource, resourceRequest plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	req := &gputypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	// if err := req.Validate(); err != nil {
	// 	return nil, err
	// }
	originResource := &gputypes.WorkloadResource{}
	if err := originResource.Parse(resource); err != nil {
		return nil, err
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		log.WithFunc("resource.gpu.CalculateRealloc").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	// put resources back into the resource pool
	nodeResourceInfo.Usage.Sub(&gputypes.NodeResource{
		GPUMap: originResource.GPUMap,
	})

	newReq := req.DeepCopy()
	newReq.MergeFromResource(originResource, req.MergeType)
	if err = newReq.Validate(); err != nil {
		return nil, err
	}

	var enginesParams []*gputypes.EngineParams
	var workloadsResource []*gputypes.WorkloadResource
	if enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, 1, newReq); err != nil {
		return nil, err
	}

	engineParams := enginesParams[0]
	newResource := workloadsResource[0]

	deltaWorkloadResource := newResource.DeepCopy()
	deltaWorkloadResource.Sub(originResource)

	return resourcetypes.RawParams{
		"engine_params":     engineParams,
		"delta_resource":    deltaWorkloadResource,
		"workload_resource": newResource,
	}, nil
}

// CalculateRemap .
func (p Plugin) CalculateRemap(context.Context, string, map[string]plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	return resourcetypes.RawParams{
		"engine_params_map": nil,
	}, nil
}

func (p Plugin) doAlloc(resourceInfo *gputypes.NodeResourceInfo, deployCount int, req *gputypes.WorkloadResourceRequest) ([]*gputypes.EngineParams, []*gputypes.WorkloadResource, error) {
	enginesParams := []*gputypes.EngineParams{}
	workloadsResource := []*gputypes.WorkloadResource{}
	var err error

	availableResource := resourceInfo.GetAvailableResource()
	for i := 0; i < deployCount; i++ {
		gpuMap := gputypes.GPUMap{}
		var addrs []string
		nMatched := 0
		for _, reqInfo := range req.GPUs {
			matched := false
			for addr, info := range availableResource.GPUMap {
				if strings.Contains(info.Product, reqInfo.Product) {
					delete(availableResource.GPUMap, addr)
					gpuMap[addr] = info
					addrs = append(addrs, addr)
					matched = true
					nMatched++
					break
				}
			}
			if !matched {
				break
			}
		}
		if nMatched == len(req.GPUs) {
			workloadsResource = append(workloadsResource, &gputypes.WorkloadResource{GPUMap: gpuMap})
			enginesParams = append(enginesParams, &gputypes.EngineParams{
				Addrs: addrs,
			})
		} else {
			err = coretypes.ErrInsufficientResource
			break
		}
	}

	return enginesParams, workloadsResource, err
}
