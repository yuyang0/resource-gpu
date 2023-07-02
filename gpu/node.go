package gpu

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/cockroachdb/errors"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"

	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	"github.com/sanity-io/litter"
	gputypes "github.com/yuyang0/resource-gpu/gpu/types"
)

const (
	maxCapacity = 1000000
)

// AddNode .
func (p Plugin) AddNode(ctx context.Context, nodename string, resource plugintypes.NodeResourceRequest, info *enginetypes.Info) (resourcetypes.RawParams, error) {
	// try to get the node resource
	var err error
	if _, err = p.doGetNodeResourceInfo(ctx, nodename); err == nil {
		return nil, coretypes.ErrNodeExists
	}

	if !errors.IsAny(err, coretypes.ErrInvaildCount, coretypes.ErrNodeNotExists) {
		log.WithFunc("resource.gpu.AddNode").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	req := &gputypes.NodeResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}
	capacity := gputypes.NewNodeResource(req.GPUMap)
	// try to fetch resource from info
	if info != nil && info.Resources != nil { //nolint
		if capacity.Len() == 0 {
			if b, ok := info.Resources[p.name]; ok {
				err := json.Unmarshal(b, capacity)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	nodeResourceInfo := &gputypes.NodeResourceInfo{
		Capacity: capacity,
		Usage:    gputypes.NewNodeResource(nil),
	}

	if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		return nil, err
	}
	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
	}, nil
}

// RemoveNode .
func (p Plugin) RemoveNode(ctx context.Context, nodename string) error {
	var err error
	if _, err = p.store.Delete(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename)); err != nil {
		log.WithFunc("resource.gpu.RemoveNode").WithField("node", nodename).Error(ctx, err, "faield to delete node")
	}
	return err
}

// GetNodesDeployCapacity returns available nodes and total capacity
func (p Plugin) GetNodesDeployCapacity(ctx context.Context, nodenames []string, resource plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.gpu.GetNodesDeployCapacity")
	req := &gputypes.WorkloadResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodesDeployCapacityMap := map[string]*plugintypes.NodeDeployCapacity{}
	total := 0

	nodesResourceInfos, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfos {
		nodeDeployCapacity := p.doGetNodeDeployCapacity(nodeResourceInfo, req)
		if nodeDeployCapacity.Capacity > 0 {
			nodesDeployCapacityMap[nodename] = nodeDeployCapacity
			if total == math.MaxInt || nodeDeployCapacity.Capacity == math.MaxInt {
				total = math.MaxInt
			} else {
				total += nodeDeployCapacity.Capacity
			}
		}
	}
	return resourcetypes.RawParams{
		"nodes_deploy_capacity_map": nodesDeployCapacityMap,
		"total":                     total,
	}, nil
}

// SetNodeResourceCapacity sets the amount of total resource info
func (p Plugin) SetNodeResourceCapacity(
	ctx context.Context, nodename string,
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	delta bool, incr bool,
) (
	resourcetypes.RawParams, error,
) {
	logger := log.WithFunc("resource.gpu.SetNodeResourceCapacity").WithField("node", "nodename")
	req, nodeResource, _, err := p.parseNodeResourceInfos(resourceRequest, resource, nil)
	if err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}

	origin := nodeResourceInfo.Capacity
	before := origin.DeepCopy()

	if !delta && req != nil {
		req.LoadFromOrigin(origin, resourceRequest)
	}
	nodeResourceInfo.Capacity = p.calculateNodeResource(req, nodeResource, origin, nil, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Capacity,
	}, nil
}

// GetNodeResourceInfo .
func (p Plugin) GetNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, _, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

// SetNodeResourceInfo .
func (p Plugin) SetNodeResourceInfo(ctx context.Context, nodename string, capacity plugintypes.NodeResource, usage plugintypes.NodeResource) error {
	capacityResource := &gputypes.NodeResource{}
	usageResource := &gputypes.NodeResource{}
	if err := capacityResource.Parse(capacity); err != nil {
		return err
	}
	if err := usageResource.Parse(usage); err != nil {
		return err
	}
	resourceInfo := &gputypes.NodeResourceInfo{
		Capacity: capacityResource,
		Usage:    usageResource,
	}

	return p.doSetNodeResourceInfo(ctx, nodename, resourceInfo)
}

// SetNodeResourceUsage .
func (p Plugin) SetNodeResourceUsage(
	ctx context.Context, nodename string,
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	workloadsResource []plugintypes.WorkloadResource,
	delta bool, incr bool,
) (
	resourcetypes.RawParams, error,
) {

	logger := log.WithFunc("resource.gpu.SetNodeResourceUsage").WithField("node", "nodename")
	req, nodeResource, wrksResource, err := p.parseNodeResourceInfos(resourceRequest, resource, workloadsResource)
	if err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}

	origin := nodeResourceInfo.Usage
	before := origin.DeepCopy()

	nodeResourceInfo.Usage = p.calculateNodeResource(req, nodeResource, origin, wrksResource, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Usage,
	}, nil
}

// GetMostIdleNode .
func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (resourcetypes.RawParams, error) {
	var mostIdleNode string
	var minIdle = math.MaxFloat64

	nodesResourceInfo, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfo {
		idle := float64(nodeResourceInfo.UsageLen()) / float64(nodeResourceInfo.CapLen())

		if idle < minIdle {
			mostIdleNode = nodename
			minIdle = idle
		}
	}
	return resourcetypes.RawParams{
		"nodename": mostIdleNode,
		"priority": priority,
	}, nil
}

// FixNodeResource .
// use workloadsReource to construct a new NodeResource, then use this NodeResource to repace Usage
func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, actuallyWorkloadsUsage, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	if len(diffs) != 0 {
		nodeResourceInfo.Usage = &gputypes.NodeResource{
			GPUMap: actuallyWorkloadsUsage.GPUMap,
		}
		if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
			log.WithFunc("resource.gpu.FixNodeResource").Error(ctx, err)
			diffs = append(diffs, err.Error())
		}
	}
	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

func (p Plugin) getNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*gputypes.NodeResourceInfo, *gputypes.WorkloadResource, []string, error) {
	logger := log.WithFunc("resource.gpu.getNodeResourceInfo").WithField("node", nodename)
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.Error(ctx, err)
		return nodeResourceInfo, nil, nil, err
	}

	actuallyWorkloadsUsage := &gputypes.WorkloadResource{GPUMap: gputypes.GPUMap{}}
	for _, workloadResource := range workloadsResource {
		workloadUsage := &gputypes.WorkloadResource{}
		if err := workloadUsage.Parse(workloadResource); err != nil {
			logger.Error(ctx, err)
			return nil, nil, nil, err
		}
		actuallyWorkloadsUsage.Add(workloadUsage)
	}

	diffs := []string{}

	if actuallyWorkloadsUsage.Len() != nodeResourceInfo.UsageLen() {
		diffs = append(diffs, fmt.Sprintf("node.GPUUsed != sum(workload.GPURequest): %.2d != %.2d", nodeResourceInfo.UsageLen(), actuallyWorkloadsUsage.Len()))
	}
	for addr := range actuallyWorkloadsUsage.GPUMap {
		if _, ok := nodeResourceInfo.Usage.GPUMap[addr]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s not in usage", addr))
		}
	}

	return nodeResourceInfo, actuallyWorkloadsUsage, diffs, nil
}

func (p Plugin) doGetNodeResourceInfo(ctx context.Context, nodename string) (*gputypes.NodeResourceInfo, error) {
	key := fmt.Sprintf(nodeResourceInfoKey, nodename)
	resp, err := p.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	r := &gputypes.NodeResourceInfo{}
	switch resp.Count {
	case 0:
		return r, errors.Wrapf(coretypes.ErrNodeNotExists, "key: %s", nodename)
	case 1:
		if err := json.Unmarshal(resp.Kvs[0].Value, r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.Wrapf(coretypes.ErrInvaildCount, "key: %s", nodename)
	}
}

func (p Plugin) doGetNodesResourceInfo(ctx context.Context, nodenames []string) (map[string]*gputypes.NodeResourceInfo, error) {
	keys := []string{}
	for _, nodename := range nodenames {
		keys = append(keys, fmt.Sprintf(nodeResourceInfoKey, nodename))
	}
	resps, err := p.store.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := map[string]*gputypes.NodeResourceInfo{}

	for _, resp := range resps {
		r := &gputypes.NodeResourceInfo{}
		if err := json.Unmarshal(resp.Value, r); err != nil {
			return nil, err
		}
		result[utils.Tail(string(resp.Key))] = r
	}
	return result, nil
}

func (p Plugin) doSetNodeResourceInfo(ctx context.Context, nodename string, resourceInfo *gputypes.NodeResourceInfo) error {
	if err := resourceInfo.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(resourceInfo)
	if err != nil {
		return err
	}

	_, err = p.store.Put(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename), string(data))
	return err
}

func (p Plugin) doGetNodeDeployCapacity(nodeResourceInfo *gputypes.NodeResourceInfo, req *gputypes.WorkloadResourceRequest) *plugintypes.NodeDeployCapacity {
	availableResource := nodeResourceInfo.GetAvailableResource()

	capacityInfo := &plugintypes.NodeDeployCapacity{
		Weight: 1, // TODO why 1?
	}
	if req.Count == 0 { //nolint
		// count equals to 0, then assign a big value to capacity
		capacityInfo.Capacity = maxCapacity
	} else {
		for {
			nMatched := 0
			for _, reqInfo := range req.GPUs {
				matched := false
				for addr, info := range availableResource.GPUMap {
					if strings.Contains(info.Product, reqInfo.Product) {
						delete(availableResource.GPUMap, addr)
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
				capacityInfo.Capacity++
			} else {
				break
			}
		}
	}
	capacityInfo.Usage = float64(nodeResourceInfo.UsageLen()) / float64(nodeResourceInfo.CapLen())
	capacityInfo.Rate = float64(len(req.GPUs)) / float64(nodeResourceInfo.CapLen())
	return capacityInfo
}

// 丢弃origin，完全用新数据重写
func (p Plugin) overwriteNodeResource(req *gputypes.NodeResourceRequest, nodeResource *gputypes.NodeResource, workloadsResource []*gputypes.WorkloadResource) *gputypes.NodeResource {
	resp := (&gputypes.NodeResource{}).DeepCopy() // init nil pointer!
	if req != nil {
		nodeResource = &gputypes.NodeResource{
			GPUMap: req.GPUMap,
		}
	}

	if nodeResource != nil {
		resp.Add(nodeResource)
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &gputypes.NodeResource{
			GPUMap: workloadResource.GPUMap,
		}
		resp.Add(nodeResource)
	}
	return resp
}

// 增量更新
func (p Plugin) incrUpdateNodeResource(req *gputypes.NodeResourceRequest, nodeResource *gputypes.NodeResource, origin *gputypes.NodeResource, workloadsResource []*gputypes.WorkloadResource, incr bool) *gputypes.NodeResource {
	resp := origin.DeepCopy()
	if req != nil {
		nodeResource = &gputypes.NodeResource{
			GPUMap: req.GPUMap,
		}
	}

	if nodeResource != nil {
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &gputypes.NodeResource{
			GPUMap: workloadResource.GPUMap,
		}
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
	}
	return resp
}

// calculateNodeResource priority: node resource request > node resource > workload resource args list
func (p Plugin) calculateNodeResource(req *gputypes.NodeResourceRequest, nodeResource *gputypes.NodeResource, origin *gputypes.NodeResource, workloadsResource []*gputypes.WorkloadResource, delta bool, incr bool) *gputypes.NodeResource {
	// req, nodeResource, workloadResource只有一个会生效, 优先级是req, nodeResource, workloadsReource
	// 如果delta为false那就不考虑origin
	// 如果delta为true那就把3者中生效的那个加到origin上
	if origin == nil || !delta { // 重写
		return p.overwriteNodeResource(req, nodeResource, workloadsResource)
	} else { //nolint
		return p.incrUpdateNodeResource(req, nodeResource, origin, workloadsResource, incr)
	}
}

func (p Plugin) parseNodeResourceInfos(
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	workloadsResource []plugintypes.WorkloadResource,
) (
	*gputypes.NodeResourceRequest,
	*gputypes.NodeResource,
	[]*gputypes.WorkloadResource,
	error,
) {
	var req *gputypes.NodeResourceRequest
	var nodeResource *gputypes.NodeResource
	wrksResource := []*gputypes.WorkloadResource{}

	if resourceRequest != nil {
		req = &gputypes.NodeResourceRequest{}
		if err := req.Parse(resourceRequest); err != nil {
			return nil, nil, nil, err
		}
	}

	if resource != nil {
		nodeResource = &gputypes.NodeResource{}
		if err := nodeResource.Parse(resource); err != nil {
			return nil, nil, nil, err
		}
	}

	for _, workloadResource := range workloadsResource {
		wrkResource := &gputypes.WorkloadResource{}
		if err := wrkResource.Parse(workloadResource); err != nil {
			return nil, nil, nil, err
		}
		wrksResource = append(wrksResource, wrkResource)
	}

	return req, nodeResource, wrksResource, nil
}
