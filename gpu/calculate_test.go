package gpu

import (
	"context"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-gpu/gpu/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

func TestCalculateDeploy(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// invalid opts
	req := plugintypes.WorkloadResourceRequest{
		"count": 1,
		"gpus": []types.GPUInfo{
			{
				Product: "3070",
			},
			{
				Product: "3090",
			},
		},
	}
	_, err := cm.CalculateDeploy(ctx, node, 100, req)
	assert.True(t, errors.Is(err, types.ErrInvalidGPU))

	// non-existent node
	req = plugintypes.WorkloadResourceRequest{
		"count": 2,
		"gpus": []types.GPUInfo{
			{
				Product: "3070",
			},
			{
				Product: "3090",
			},
		},
	}
	_, err = cm.CalculateDeploy(ctx, "xxx", 100, req)
	assert.True(t, errors.Is(err, coretypes.ErrNodeNotExists))

	// normal cases
	// 1. empty request
	d, err := cm.CalculateDeploy(ctx, node, 4, nil)
	assert.Nil(t, err)
	assert.NotNil(t, d["engines_params"])
	eParams := d["engines_params"].([]*types.EngineParams)
	wResources := d["workloads_resource"].([]*types.WorkloadResource)
	assert.Len(t, eParams, 4)
	assert.Len(t, wResources, 4)
	for i := 0; i < 4; i++ {
		assert.Len(t, eParams[i].Addrs, 0)
		assert.Len(t, wResources[i].GPUMap, 0)
	}
	// has enough resource
	d, err = cm.CalculateDeploy(ctx, node, 4, req)
	assert.Nil(t, err)
	assert.NotNil(t, d["engines_params"])
	eParams = d["engines_params"].([]*types.EngineParams)
	assert.Len(t, eParams, 4)

	// don't have enough resource
	d, err = cm.CalculateDeploy(ctx, node, 5, req)
	assert.Error(t, err)
}

func TestCalculateRealloc(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// set capacity
	resource := plugintypes.NodeResource{
		"gpu_map": types.GPUMap{
			"0000:02:00.0": types.GPUInfo{
				Address: "0000:02:00.0",
				Product: "GA104 [GeForce RTX 3070]",
				Vendor:  "NVIDIA Corporation",
			},
			"0000:82:00.0": types.GPUInfo{
				Address: "0000:82:00.0",
				Product: "GA105 [GeForce RTX 3090]",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}

	_, err := cm.SetNodeResourceUsage(ctx, node, resource, nil, nil, false, true)
	assert.Nil(t, err)

	origin := plugintypes.WorkloadResource{}
	req := plugintypes.WorkloadResourceRequest{}

	// non-existent node
	_, err = cm.CalculateRealloc(ctx, "xxx", origin, req)
	assert.True(t, errors.Is(err, coretypes.ErrNodeNotExists))

	// normal cases
	// 1. empty request and resource
	d, err := cm.CalculateRealloc(ctx, node, nil, nil)
	assert.Nil(t, err)
	eParams := d["engine_params"].(*types.EngineParams)
	wResource := d["workload_resource"].(*types.WorkloadResource)
	assert.Len(t, eParams.Addrs, 0)
	assert.Len(t, wResource.GPUMap, 0)
	// 2. empty request
	origin = plugintypes.WorkloadResource{
		"gpu_map": types.GPUMap{
			"0000:82:00.0": types.GPUInfo{
				Product: "GA105 [GeForce RTX 3090]",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, nil)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	assert.Len(t, eParams.Addrs, 1)
	assert.Len(t, wResource.GPUMap, 1)
	assert.Equal(t, eParams.Addrs[0], maps.Keys(wResource.GPUMap)[0])
	// 3. overwirte resource with request
	origin = plugintypes.WorkloadResource{
		"gpu_map": types.GPUMap{
			"0000:82:00.0": types.GPUInfo{
				Product: "GA105 [GeForce RTX 3090]",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}

	req = plugintypes.WorkloadResourceRequest{
		"merge_type": types.MergeTotol,
		"count":      2,
		"gpus": []types.GPUInfo{
			{
				Product: "GA105 [GeForce RTX 3090]",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	assert.Len(t, eParams.Addrs, 2)
	assert.True(t, strings.HasPrefix(eParams.Addrs[0], "0000:8"))
	assert.True(t, strings.HasPrefix(eParams.Addrs[1], "0000:8"))
	assert.Len(t, wResource.GPUMap, 2)
	for _, info := range maps.Values(wResource.GPUMap) {
		assert.True(t, strings.Contains(info.Product, "3090"))
	}
	// 4. Add origin resources to request
	req = plugintypes.WorkloadResourceRequest{
		"merge_type": types.MergeAdd,
		"count":      2,
		"gpus": []types.GPUInfo{
			{
				Product: "3090",
				Vendor:  "NVIDIA Corporation",
			},
			{
				Product: "3070",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	assert.Len(t, eParams.Addrs, 3)
	// assert.True(t, strings.Contains(eParams.Addrs[0], "8"))
	// assert.True(t, strings.Contains(eParams.Addrs[1], "8"))
	assert.Len(t, wResource.GPUMap, 3)
	// for _, info := range maps.Values(wResource.GPUMap) {
	// 	assert.True(t, strings.Contains(info.Product, "3090"))
	// }

	// remove GPU
	req = plugintypes.WorkloadResourceRequest{
		"merge_type": types.MergeSub,
		"count":      2,
		"gpus": []types.GPUInfo{
			{
				Product: "3090",
				Vendor:  "NVIDIA Corporation",
			},
			{
				Product: "3070",
				Vendor:  "NVIDIA Corporation",
			},
		},
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	assert.Len(t, eParams.Addrs, 1)
	assert.True(t, strings.HasPrefix(eParams.Addrs[0], "0000:0"))
	assert.Len(t, wResource.GPUMap, 1)
	for _, info := range maps.Values(wResource.GPUMap) {
		assert.True(t, strings.Contains(info.Product, "3070"))
	}
}

func TestCalculateRemap(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]
	d, err := cm.CalculateRemap(ctx, node, nil)

	assert.NoError(t, err)
	assert.Nil(t, d["engine_params_map"])

}
