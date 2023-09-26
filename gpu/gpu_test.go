package gpu

import (
	"context"
	"fmt"
	"testing"

	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/yuyang0/resource-gpu/gpu/types"
)

func TestName(t *testing.T) {
	cm := initGPU(context.Background(), t)
	assert.Equal(t, cm.name, cm.Name())
}

func initGPU(ctx context.Context, t *testing.T) *Plugin {
	config := coretypes.Config{
		Etcd: coretypes.EtcdConfig{
			Prefix: "/gpu",
		},
		Scheduler: coretypes.SchedulerConfig{
			MaxShare:  -1,
			ShareBase: 100,
		},
	}

	cm, err := NewPlugin(ctx, config, t)
	assert.NoError(t, err)
	return cm
}

func generateNodes(
	ctx context.Context, t *testing.T, cm *Plugin,
	nums int, index int,
) []string {
	reqs := generateNodeResourceRequests(t, nums, index, "test", 8)
	info := &enginetypes.Info{NCPU: 8, MemTotal: 2048}
	names := []string{}
	for name, req := range reqs {
		_, err := cm.AddNode(ctx, name, req, info)
		assert.NoError(t, err)
		names = append(names, name)
	}
	t.Cleanup(func() {
		for name := range reqs {
			assert.NoError(t, cm.RemoveNode(ctx, name))
		}
	})
	return names
}

func generateEmptyNodes(
	ctx context.Context, t *testing.T, cm *Plugin,
	nums int, index int,
) []string {
	reqs := generateNodeResourceRequests(t, nums, index, "test-empty", 0)
	info := &enginetypes.Info{NCPU: 8, MemTotal: 2048}
	names := []string{}
	for name, req := range reqs {
		_, err := cm.AddNode(ctx, name, req, info)
		assert.NoError(t, err)
		names = append(names, name)
	}
	t.Cleanup(func() {
		for name := range reqs {
			assert.NoError(t, cm.RemoveNode(ctx, name))
		}
	})
	return names
}

func generateNodeResourceRequests(t *testing.T, nums int, index int, namePrefix string, numGPUs int) map[string]plugintypes.NodeResourceRequest {
	gpuMap := types.ProdCountMap{
		"nvidia-3070": numGPUs / 2,
		"nvidia-3090": numGPUs / 2,
	}

	for prod, count := range gpuMap {
		if count <= 0 {
			delete(gpuMap, prod)
		}
	}

	infos := map[string]plugintypes.NodeResourceRequest{}
	for i := index; i < index+nums; i++ {
		info := plugintypes.NodeResourceRequest{
			"prod_count_map": gpuMap,
		}
		infos[fmt.Sprintf("%s%v", namePrefix, i)] = info
	}
	return infos
}
