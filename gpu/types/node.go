package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// NodeResource indicate node cpumem resource
type NodeResource struct {
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

func NewNodeResource(gm ProdCountMap) *NodeResource {
	r := &NodeResource{
		ProdCountMap: gm,
	}
	if r.ProdCountMap == nil {
		r.ProdCountMap = ProdCountMap{}
	}
	return r
}

func (r *NodeResource) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"prod_count_map": r.ProdCountMap,
	}
}

// Parse .
func (r *NodeResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, r)
}

func (r *NodeResource) Validate() error {
	return r.ProdCountMap.Validate()
}

// DeepCopy .
func (r *NodeResource) DeepCopy() *NodeResource {
	res := &NodeResource{
		ProdCountMap: r.ProdCountMap.DeepCopy(),
	}
	return res
}

// Add .
func (r *NodeResource) Add(r1 *NodeResource) {
	r.ProdCountMap.Add(r1.ProdCountMap)
}

// Sub .
func (r *NodeResource) Sub(r1 *NodeResource) {
	r.ProdCountMap.Sub(r1.ProdCountMap)
}

// Count
func (r *NodeResource) Count() int {
	return r.ProdCountMap.TotalCount()
}

// NodeResourceInfo indicate cpumem capacity and usage
type NodeResourceInfo struct {
	Capacity *NodeResource `json:"capacity"`
	Usage    *NodeResource `json:"usage"`
}

func (n *NodeResourceInfo) CapCount() int {
	return n.Capacity.Count()
}

func (n *NodeResourceInfo) UsageCount() int {
	return n.Usage.Count()
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
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

func (n *NodeResourceRequest) Parse(rawParams resourcetypes.RawParams) error {
	if err := mapstructure.Decode(rawParams, n); err != nil {
		return err
	}
	if n.ProdCountMap == nil {
		n.ProdCountMap = ProdCountMap{}
	}
	return nil
}

func (n *NodeResourceRequest) Validate() error {
	return n.ProdCountMap.Validate()
}

func (n *NodeResourceRequest) Count() int {
	return n.ProdCountMap.TotalCount()
}

// Merge fields to NodeResourceRequest.
func (n *NodeResourceRequest) LoadFromOrigin(nodeResource *NodeResource, resourceRequest resourcetypes.RawParams) {
	if n == nil {
		return
	}
	if !resourceRequest.IsSet("prod_count_map") {
		n.ProdCountMap = nodeResource.ProdCountMap
	}
}
