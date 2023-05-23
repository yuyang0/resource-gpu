package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// NodeResource indicate node cpumem resource
type NodeResource struct {
	GPUMap GPUMap `json:"gpu_map" mapstructure:"gpu_map"`
}

func NewNodeResource(gm GPUMap) *NodeResource {
	r := &NodeResource{
		GPUMap: gm,
	}
	if r.GPUMap == nil {
		r.GPUMap = GPUMap{}
	}
	return r
}

// Parse .
func (r *NodeResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, r)
}

func (r *NodeResource) Validate() error {
	return r.GPUMap.Validate()
}

// DeepCopy .
func (r *NodeResource) DeepCopy() *NodeResource {
	res := &NodeResource{
		GPUMap: GPUMap{},
	}

	for addr := range r.GPUMap {
		res.GPUMap[addr] = r.GPUMap[addr]
	}
	return res
}

// Add .
func (r *NodeResource) Add(r1 *NodeResource) {
	for addr, info := range r1.GPUMap {
		r.GPUMap[addr] = info
	}
}

// Sub .
func (r *NodeResource) Sub(r1 *NodeResource) {
	for addr := range r1.GPUMap {
		delete(r.GPUMap, addr)
	}
}

// Len
func (r *NodeResource) Len() int {
	return len(r.GPUMap)
}

// NodeResourceInfo indicate cpumem capacity and usage
type NodeResourceInfo struct {
	Capacity *NodeResource `json:"capacity"`
	Usage    *NodeResource `json:"usage"`
}

func (n *NodeResourceInfo) CapLen() int {
	return len(n.Capacity.GPUMap)
}
func (n *NodeResourceInfo) UsageLen() int {
	return len(n.Usage.GPUMap)
}

// DeepCopy .
func (n *NodeResourceInfo) DeepCopy() *NodeResourceInfo {
	return &NodeResourceInfo{
		Capacity: n.Capacity.DeepCopy(),
		Usage:    n.Usage.DeepCopy(),
	}
}

func (n *NodeResourceInfo) Validate() error {
	if err := n.Capacity.Validate(); err != nil {
		return err
	}
	return n.Usage.Validate()
}

func (n *NodeResourceInfo) GetAvailableResource() *NodeResource {
	availableResource := n.Capacity.DeepCopy()
	availableResource.Sub(n.Usage)

	return availableResource
}

// NodeResourceRequest includes all possible fields passed by eru-core for editing node, it not parsed!
type NodeResourceRequest struct {
	GPUMap GPUMap `json:"gpu_map" mapstructure:"gpu_map"`
}

func (n *NodeResourceRequest) Parse(rawParams resourcetypes.RawParams) error {
	if err := mapstructure.Decode(rawParams, n); err != nil {
		return err
	}
	if n.GPUMap == nil {
		n.GPUMap = GPUMap{}
	}
	return nil
}

// Merge fields to NodeResourceRequest.
func (n *NodeResourceRequest) LoadFromOrigin(nodeResource *NodeResource, resourceRequest resourcetypes.RawParams) {
	if n == nil {
		return
	}
	if !resourceRequest.IsSet("gpu_map") {
		n.GPUMap = nodeResource.GPUMap
	}
}
