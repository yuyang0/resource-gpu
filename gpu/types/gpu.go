package types

import (
	"encoding/json"

	resourcetypes "github.com/projecteru2/core/resource/types"
)

type GPUInfo struct {
	Address string `json:"address" mapstructure:"address"`
	Index   int    `json:"index" mapstructure:"index"`
	// example value: "NVIDIA Corporation"
	Vendor string `json:"vendor" mapstructure:"vendor"`
	// example value: "GA104 [GeForce RTX 3070]"
	Product string `json:"product" mapstructure:"product"`

	// NUMA NUMAInfo
	NumaID string `json:"numa_id" mapstructure:"numa_id"`

	// Cores   int   `json:"cores" mapstructure:"cores"`
	GMemory int64 `json:"gmemory" mapstructure:"gmemory"`
}

type GPUMap map[string]GPUInfo

func (gm GPUMap) Load(rawParams resourcetypes.RawParams) error {
	for k, v := range rawParams {
		g := GPUInfo{}
		vs := v.(string)
		if err := json.Unmarshal([]byte(vs), &g); err != nil {
			return err
		}
		gm[k] = g
	}
	return nil
}

func (gm GPUMap) Add(g1 GPUMap) {
	for addr, info := range g1 {
		gm[addr] = info
	}
}

func (gm GPUMap) Sub(g1 GPUMap) {
	for addr := range g1 {
		delete(gm, addr)
	}
}

// NUMA map[address]nodeID
type NUMA map[string]string
