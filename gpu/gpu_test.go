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
	reqs := generateNodeResourceRequests(t, nums, index)
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

func generateNodeResourceRequests(t *testing.T, nums int, index int) map[string]plugintypes.NodeResourceRequest {
	addrs := []string{
		"0000:00:00.0", "0000:01:00.0", "0000:02:00.0", "0000:03:00.0",
		"0000:80:00.0", "0000:81:00.0", "0000:82:00.0", "0000:83:00.0",
	}
	gInfo := types.GPUInfo{
		Product: "GA104 [GeForce RTX 3070]",
		Vendor:  "NVIDIA Corporation",
	}
	gpuMap := types.GPUMap{}
	for idx, addr := range addrs {
		gInfo.Address = addr
		if idx >= len(addrs)/2 {
			gInfo.Product = "GA105 [GeForce RTX 3090]"
		}
		gpuMap[addr] = gInfo
	}

	infos := map[string]plugintypes.NodeResourceRequest{}
	for i := index; i < index+nums; i++ {
		info := plugintypes.NodeResourceRequest{
			"gpu_map": gpuMap,
		}
		infos[fmt.Sprintf("test%v", i)] = info
	}
	return infos
}
