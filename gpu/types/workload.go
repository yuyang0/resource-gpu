package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate GPU workload resource
type WorkloadResource struct {
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

func (w *WorkloadResource) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"prod_count_map": w.ProdCountMap,
	}
}
func (w *WorkloadResource) Validate() error {
	return w.ProdCountMap.Validate()
}

// ParseFromRawParams .
func (w *WorkloadResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, w)
}

// DeepCopy .
func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	res := &WorkloadResource{
		ProdCountMap: w.ProdCountMap.DeepCopy(),
	}
	return res
}

// Add .
func (w *WorkloadResource) Add(w1 *WorkloadResource) {
	w.ProdCountMap.Add(w1.ProdCountMap)
}

// Sub .
func (w *WorkloadResource) Sub(w1 *WorkloadResource) {
	w.ProdCountMap.Sub(w1.ProdCountMap)
}

// Count
func (w *WorkloadResource) Count() int {
	return w.ProdCountMap.TotalCount()
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

// Validate .
func (w *WorkloadResourceRequest) ValidateProd() error {
	// empty ProdCountMap means this request doesn't need GPU
	// in order to support realloc, the count can be negative, so only validate prod here
	return w.ProdCountMap.ValidateProd()
}

func (w *WorkloadResourceRequest) Validate() error {
	// empty ProdCountMap means this request doesn't need GPU
	return w.ProdCountMap.Validate()
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	return mapstructure.Decode(rawParams, w)
}

func (w *WorkloadResourceRequest) MergeFromResource(r *WorkloadResource) {
	w.ProdCountMap.Add(r.ProdCountMap)
	newMap := ProdCountMap{}
	for prod, count := range w.ProdCountMap {
		if count > 0 {
			newMap[prod] = count
		}
	}
	w.ProdCountMap = newMap
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	return &WorkloadResourceRequest{
		ProdCountMap: w.ProdCountMap.DeepCopy(),
	}
}

func (w *WorkloadResourceRequest) Count() int {
	return w.ProdCountMap.TotalCount()
}
