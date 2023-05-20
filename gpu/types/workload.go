package types

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate GPU workload resource
type WorkloadResource struct {
	GPUMap GPUMap `json:"gpu_map" mapstructure:"gpu_map"`
}

// ParseFromRawParams .
func (w *WorkloadResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, w)
}

// DeepCopy .
func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	res := &WorkloadResource{
		GPUMap: GPUMap{},
	}

	for addr, info := range w.GPUMap {
		res.GPUMap[addr] = info
	}

	return res
}

// Add .
func (w *WorkloadResource) Add(w1 *WorkloadResource) {
	w.GPUMap.Add(w1.GPUMap)
}

// Sub .
func (w *WorkloadResource) Sub(w1 *WorkloadResource) {
	w.GPUMap.Sub(w1.GPUMap)
}

// Len
func (w *WorkloadResource) Len() int {
	return len(w.GPUMap)
}

type MergeType int

const (
	MergeAdd MergeType = iota
	MergeSub
	MergeTotol
)

func (mt MergeType) Validate() bool {
	return mt >= MergeAdd && mt <= MergeTotol
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	MergeType MergeType `json:"merge_type" mapstructure:"merge_type"`
	Count     int       `json:"count" mapstructure:"count"`
	GPUs      []GPUInfo `json:"gpus" mapstructure:"gpus"`
}

// Validate .
func (w *WorkloadResourceRequest) Validate() error {
	if w.Count <= 0 || w.Count != len(w.GPUs) {
		return ErrInvalidGPU
	}
	if !w.MergeType.Validate() {
		return ErrInvalidMergeType
	}
	return nil
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	err = mapstructure.Decode(rawParams, w)
	if err != nil {
		return err
	}
	if w.Count > 1 && len(w.GPUs) == 1 {
		g := &w.GPUs[0]
		for i := 0; i < w.Count-1; i++ {
			w.GPUs = append(w.GPUs, *g)
		}
	}
	return nil
}

func (w *WorkloadResourceRequest) MergeFromResource(r *WorkloadResource, mergeTy MergeType) {
	switch mergeTy {
	case MergeAdd:
		for _, info := range r.GPUMap {
			w.GPUs = append(w.GPUs, info)
			w.Count++
		}
	case MergeSub:
		// remove gpus which matches the gpu in resources
		count := w.Count
		for _, info := range r.GPUMap {
			for idx := 0; idx < count; idx++ {
				info2 := w.GPUs[idx]
				if strings.Contains(info.Product, info2.Product) {
					w.GPUs[idx], w.GPUs[count-1] = w.GPUs[count-1], w.GPUs[idx]
					count--
					break
				}
			}
		}
		w.GPUs = w.GPUs[:count]
		w.Count = count
	case MergeTotol:
		// use request overwrite resource, so do nothing here
		return
	}
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	newReq := &WorkloadResourceRequest{
		Count: w.Count,
		GPUs:  make([]GPUInfo, len(w.GPUs)),
	}
	copy(newReq.GPUs, w.GPUs)
	return newReq
}
