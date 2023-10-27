package gpu

import (
	"context"
	"errors"
	"testing"

	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/yuyang0/resource-gpu/gpu/types"
)

func TestCalculateDeploy(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// invalid opts
	req := plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": -1,
			"nvidia-3090": 1,
		},
	}
	_, err := cm.CalculateDeploy(ctx, node, 100, req)
	assert.True(t, errors.Is(err, types.ErrInvalidGPUMap))

	// non-existent node
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
			"nvidia-3090": 1,
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
		assert.Equal(t, eParams[i].ProdCountMap.TotalCount(), 0)
		assert.Equal(t, wResources[i].Count(), 0)
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
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
			"nvidia-3090": 1,
		},
	}

	_, err := cm.SetNodeResourceUsage(ctx, node, nil, resource, nil, false, true)
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
	dResource := d["delta_resource"].(*types.WorkloadResource)
	assert.Equal(t, eParams.Count(), 0)
	assert.Equal(t, wResource.Count(), 0)
	assert.Equal(t, dResource.Count(), 0)
	// 2. empty request
	origin = plugintypes.WorkloadResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
		},
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, nil)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	dResource = d["delta_resource"].(*types.WorkloadResource)

	assert.Equal(t, eParams.Count(), 1)
	count, ok := eParams.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)

	assert.Equal(t, wResource.Count(), 1)
	count, ok = wResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)

	assert.Equal(t, dResource.Count(), 0)
	// 3. overwirte resource with request
	origin = plugintypes.WorkloadResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
		},
	}

	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 2,
		},
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	dResource = d["delta_resource"].(*types.WorkloadResource)
	assert.Equal(t, eParams.Count(), 3)
	assert.Equal(t, wResource.Count(), 3)
	assert.Equal(t, dResource.Count(), 2)

	count, ok = wResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 3)

	count, ok = dResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 2)
	// 4. Add origin resources to request
	origin = plugintypes.WorkloadResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
		},
	}
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
			"nvidia-3070": 1,
		},
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	dResource = d["delta_resource"].(*types.WorkloadResource)

	assert.Equal(t, eParams.Count(), 3)
	count, ok = eParams.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)
	count, ok = eParams.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 2)

	assert.Equal(t, wResource.Count(), 3)
	count, ok = wResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)
	count, ok = wResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 2)

	assert.Equal(t, dResource.Count(), 2)
	count, ok = dResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)
	count, ok = dResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)

	// remove GPU
	origin = plugintypes.WorkloadResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
		},
	}
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": -1,
			"nvidia-3070": 1,
		},
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	dResource = d["delta_resource"].(*types.WorkloadResource)

	assert.Equal(t, eParams.Count(), 1)
	assert.Equal(t, wResource.Count(), 1)
	count, ok = wResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)

	assert.Equal(t, dResource.Count(), 0)
	count, ok = dResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)
	count, ok = dResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, -1)
	// smaller negative count
	origin = plugintypes.WorkloadResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": 1,
		},
	}
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3090": -5,
			"nvidia-3070": 1,
		},
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams = d["engine_params"].(*types.EngineParams)
	wResource = d["workload_resource"].(*types.WorkloadResource)
	dResource = d["delta_resource"].(*types.WorkloadResource)

	assert.Equal(t, eParams.Count(), 1)
	assert.Equal(t, wResource.Count(), 1)
	count, ok = wResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)

	assert.Equal(t, dResource.Count(), 0)
	count, ok = dResource.ProdCountMap["nvidia-3070"]
	assert.True(t, ok)
	assert.Equal(t, count, 1)
	count, ok = dResource.ProdCountMap["nvidia-3090"]
	assert.True(t, ok)
	assert.Equal(t, count, -1)
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
