package gpu

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/docker/go-units"
	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/yuyang0/resource-gpu/gpu/types"
)

func TestAddNode(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]
	nodeForAdd := "test2"

	req := plugintypes.NodeResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 2,
		},
	}

	info := &enginetypes.Info{NCPU: 2, MemTotal: 4 * units.GB}

	// existent node
	_, err := cm.AddNode(ctx, node, req, info)
	assert.Equal(t, err, coretypes.ErrNodeExists)

	// normal case
	r, err := cm.AddNode(ctx, "xxx", nil, nil)
	assert.Nil(t, err)
	// check empty capacity
	nr, err := cm.GetNodeResourceInfo(ctx, "xxx", nil)
	assert.Nil(t, err)
	cv, ok := nr["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, cv.Count(), 0)
	assert.NotNil(t, cv.ProdCountMap)
	cm.RemoveNode(ctx, "xxx")

	r, err = cm.AddNode(ctx, nodeForAdd, req, info)
	assert.Nil(t, err)
	ni, ok := r["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, ni.Count(), 2)

	// test engine info
	nRes := types.NodeResource{
		ProdCountMap: types.ProdCountMap{
			"nvidia-3070": 2,
		},
	}
	data, err := json.Marshal(&nRes)
	assert.Nil(t, err)
	eInfo := &enginetypes.Info{
		Resources: map[string][]byte{
			"gpu": data,
		},
	}
	r, err = cm.AddNode(ctx, "xxx1", nil, eInfo)
	assert.Nil(t, err)

	nr, err = cm.GetNodeResourceInfo(ctx, "xxx1", nil)
	assert.Nil(t, err)
	cv, ok = nr["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, cv.Count(), 2)
	assert.NotNil(t, cv.ProdCountMap)
	cm.RemoveNode(ctx, "xxx1")
}

func TestRemoveNode(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]
	nodeForDel := "test2"

	// node which doesn't exist in store
	err := cm.RemoveNode(ctx, "xxx")
	assert.Nil(t, err)

	err = cm.RemoveNode(ctx, node)
	assert.Nil(t, err)
	err = cm.RemoveNode(ctx, nodeForDel)
	assert.Nil(t, err)

}

func TestGetNodesDeployCapacity(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateEmptyNodes(ctx, t, cm, 2, 0)
	r, err := cm.GetNodesDeployCapacity(ctx, nodes, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2*maxCapacity, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, maxCapacity, cap.Capacity)
		assert.Equal(t, float64(0), cap.Usage)
		assert.Equal(t, float64(0), cap.Rate)
	}

	nodes = generateNodes(ctx, t, cm, 2, 0)

	req := plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 2,
		},
	}

	// non-existent node
	_, err = cm.GetNodesDeployCapacity(ctx, []string{"xxx"}, req)
	assert.True(t, errors.Is(err, coretypes.ErrInvaildCount))

	// normal
	// 1. empty request
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2*maxCapacity, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, maxCapacity, cap.Capacity)
	}

	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 4, r["total"])

	// more gpu
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 3,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 2, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, 1, cap.Capacity)
	}

	// more gpu
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 5,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 0, r["total"])
	assert.Len(t, r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity), 0)

	// 2 diffirent type of gpus
	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
			"nvidia-3090": 1,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 8, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, 4, cap.Capacity)
	}

	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
			"nvidia-3090": 2,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 4, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, 2, cap.Capacity)
	}

	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 2,
			"nvidia-3090": 2,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 4, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, 2, cap.Capacity)
	}

	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 4,
			"nvidia-3090": 4,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 2, r["total"])
	for _, node := range nodes {
		cap := r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity)[node]
		assert.Equal(t, 1, cap.Capacity)
	}

	req = plugintypes.WorkloadResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 5,
			"nvidia-3090": 4,
		},
	}
	r, err = cm.GetNodesDeployCapacity(ctx, nodes, req)
	assert.Nil(t, err)
	assert.Equal(t, 0, r["total"])
	assert.Len(t, r["nodes_deploy_capacity_map"].(map[string]*plugintypes.NodeDeployCapacity), 0)
}

func TestSetNodeResourceCapacity(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	_, err := cm.GetNodeResourceInfo(ctx, node, nil)
	assert.Nil(t, err)

	r, err := cm.GetNodeResourceInfo(ctx, node, nil)
	assert.Nil(t, err)
	v, ok := r["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 8)

	nodeResource := plugintypes.NodeResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
		},
	}

	nodeResourceRequest := plugintypes.NodeResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
		},
	}

	r, err = cm.SetNodeResourceCapacity(ctx, node, nil, nil, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 8)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nil, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 8)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nodeResourceRequest, nil, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 9)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nodeResourceRequest, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 8)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nil, nodeResource, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 9)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nodeResource, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 8)

	// overwirte node resource
	r, err = cm.SetNodeResourceCapacity(ctx, node, nodeResourceRequest, nil, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nil, nodeResource, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nodeResourceRequest, nodeResource, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceCapacity(ctx, node, nil, nil, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)
}

func TestGetAndFixNodeResourceInfo(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// invalid node
	_, err := cm.GetNodeResourceInfo(ctx, "xxx", nil)
	assert.True(t, errors.Is(err, coretypes.ErrNodeNotExists))

	r, err := cm.GetNodeResourceInfo(ctx, node, nil)
	assert.Nil(t, err)
	assert.Len(t, r["diffs"].([]string), 0)
	// r.Capacity["numa"] = types.NUMA{"0": "0", "1": "1"}
	// r.Capacity["numa_memory"] = types.NUMAMemory{"0": units.GB, "1": units.GB}

	// _, err = cm.SetNodeResourceInfo(ctx, node, r.Capacity, r.Usage)
	// assert.Nil(t, err)

	workloadsResource := []plugintypes.WorkloadResource{
		{
			"prod_count_map": types.ProdCountMap{
				"nvidia-3070": 1,
				"nvidia-3090": 1,
			},
		},
	}
	r, err = cm.GetNodeResourceInfo(ctx, node, workloadsResource)
	assert.Nil(t, err)
	assert.Len(t, r["diffs"].([]string), 3)

	r, err = cm.FixNodeResource(ctx, node, workloadsResource)
	assert.Nil(t, err)
	assert.Len(t, r["diffs"].([]string), 3)
	usage := r["usage"].(*types.NodeResource)
	assert.Equal(t, usage.Count(), 2)
	// _, ok := usage.ProdCountMap["0000:81:00.0"]
	// assert.True(t, ok)
}

func TestSetNodeResourceInfo(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	r, err := cm.GetNodeResourceInfo(ctx, node, nil)
	assert.Nil(t, err)
	cv, ok := r["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, 8, cv.Count())
	uv, ok := r["usage"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, 0, uv.Count())

	rcv := resourcetypes.RawParams{
		"prod_count_map": cv.ProdCountMap,
	}
	ucv := resourcetypes.RawParams{
		"prod_count_map": uv.ProdCountMap,
	}
	err = cm.SetNodeResourceInfo(ctx, "node-2", rcv, ucv)
	assert.Nil(t, err)
}

func TestSetNodeResourceUsage(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	r, err := cm.GetNodeResourceInfo(ctx, node, nil)
	assert.Nil(t, err)
	v, ok := r["usage"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	nodeResource := plugintypes.NodeResource{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
		},
	}

	nodeResourceRequest := plugintypes.NodeResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
		},
	}

	workloadsResource := []plugintypes.WorkloadResource{
		{
			"prod_count_map": types.ProdCountMap{
				"nvidia-3070": 1,
			},
		},
	}

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, nil, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nil, nil, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nil, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nodeResource, nil, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nodeResource, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, workloadsResource, true, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, workloadsResource, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, nil, true, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 0)

	// overwirte usage node resource
	// one params
	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nil, nil, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nodeResource, nil, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nil, workloadsResource, false, false)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	// two parmas
	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nodeResource, nil, false, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nil, workloadsResource, false, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	r, err = cm.SetNodeResourceUsage(ctx, node, nil, nodeResource, workloadsResource, false, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)

	// three params
	r, err = cm.SetNodeResourceUsage(ctx, node, nodeResourceRequest, nodeResource, workloadsResource, false, true)
	assert.Nil(t, err)
	v, ok = r["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Count(), 1)
}

func TestGetMostIdleNode(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	nodes := generateNodes(ctx, t, cm, 2, 0)

	usage := plugintypes.NodeResourceRequest{
		"prod_count_map": types.ProdCountMap{
			"nvidia-3070": 1,
		},
	}

	_, err := cm.SetNodeResourceUsage(ctx, nodes[1], nil, usage, nil, false, false)
	assert.Nil(t, err)

	r, err := cm.GetMostIdleNode(ctx, nodes)
	assert.Nil(t, err)
	assert.Equal(t, r["nodename"].(string), nodes[0])

	nodes = append(nodes, "node-x")
	_, err = cm.GetMostIdleNode(ctx, nodes)
	assert.Error(t, err)
}
