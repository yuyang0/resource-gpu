package gpu

import (
	"context"

	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	gputypes "github.com/yuyang0/resource-gpu/gpu/types"
)

// CalculateDeploy .
func (p Plugin) CalculateDeploy(
	ctx context.Context, nodename string, deployCount int,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateDeployResponse, error,
) {
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

	epRaws := make([]resourcetypes.RawParams, 0, len(enginesParams))
	for _, ep := range enginesParams {
		epRaws = append(epRaws, ep.AsRawParams())
	}
	wrRaws := make([]resourcetypes.RawParams, 0, len(workloadsResource))
	for _, wr := range workloadsResource {
		wrRaws = append(wrRaws, wr.AsRawParams())
	}
	return &plugintypes.CalculateDeployResponse{
		EnginesParams:     epRaws,
		WorkloadsResource: wrRaws,
	}, nil
}

// CalculateRealloc .
func (p Plugin) CalculateRealloc(
	ctx context.Context, nodename string,
	resource plugintypes.WorkloadResource,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateReallocResponse, error,
) {
	req := &gputypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	// realloc needs negative count, so only validate prod here.
	if err := req.ValidateProd(); err != nil {
		return nil, err
	}
	originResource := &gputypes.WorkloadResource{}
	if err := originResource.Parse(resource); err != nil {
		return nil, err
	}
	if err := originResource.Validate(); err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		log.WithFunc("resource.gpu.CalculateRealloc").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	// put resources back into the resource pool
	nodeResourceInfo.Usage.Sub(&gputypes.NodeResource{
		ProdCountMap: originResource.ProdCountMap,
	})

	newReq := req.DeepCopy()
	newReq.MergeFromResource(originResource)

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

	return &plugintypes.CalculateReallocResponse{
		EngineParams:     engineParams.AsRawParams(),
		DeltaResource:    deltaWorkloadResource.AsRawParams(),
		WorkloadResource: newResource.AsRawParams(),
	}, nil
}

// CalculateRemap .
func (p Plugin) CalculateRemap(context.Context, string, map[string]plugintypes.WorkloadResource) (*plugintypes.CalculateRemapResponse, error) {
	return &plugintypes.CalculateRemapResponse{
		EngineParamsMap: nil,
	}, nil
}

func (p Plugin) doAlloc(resourceInfo *gputypes.NodeResourceInfo, deployCount int, req *gputypes.WorkloadResourceRequest) ([]*gputypes.EngineParams, []*gputypes.WorkloadResource, error) {
	enginesParams := []*gputypes.EngineParams{}
	workloadsResource := []*gputypes.WorkloadResource{}
	var err error

	availableResource := resourceInfo.GetAvailableResource()
	for i := 0; i < deployCount; i++ {
		prodCountMap := gputypes.ProdCountMap{}
		for reqProd, reqCount := range req.ProdCountMap {
			capCount, ok := availableResource.ProdCountMap[reqProd]
			if !ok || capCount < reqCount {
				err = coretypes.ErrInsufficientResource
				return enginesParams, workloadsResource, err
			}
			availableResource.ProdCountMap[reqProd] -= reqCount
			prodCountMap[reqProd] = reqCount
		}
		if req.Count() == prodCountMap.TotalCount() {
			workloadsResource = append(workloadsResource, &gputypes.WorkloadResource{
				ProdCountMap: prodCountMap.DeepCopy(),
			})
			enginesParams = append(enginesParams, &gputypes.EngineParams{
				ProdCountMap: prodCountMap.DeepCopy(),
			})
		} else {
			err = coretypes.ErrInsufficientResource
			break
		}
	}
	return enginesParams, workloadsResource, err
}
