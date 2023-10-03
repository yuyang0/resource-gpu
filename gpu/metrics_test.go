package gpu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsDescription(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	md, err := cm.GetMetricsDescription(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, md)
	assert.Len(t, *md, 2)
}

func TestGetMetrics(t *testing.T) {
	ctx := context.Background()
	cm := initGPU(ctx, t)
	_, err := cm.GetMetrics(ctx, "", "")
	assert.Error(t, err)

	nodes := generateNodes(ctx, t, cm, 1, -1)
	resp, err := cm.GetMetrics(ctx, "testpod", nodes[0])
	assert.NoError(t, err)
	for _, mt := range *resp {
		assert.Len(t, mt.Labels, 3)
		assert.Equal(t, mt.Labels[0], "testpod")
		assert.True(t, mt.Labels[2] == "nvidia-3070" || mt.Labels[2] == "nvidia-3090")
		switch mt.Name {
		case "gpu_capacity":
			assert.Equal(t, mt.Value, "4")
		case "gpu_used":
			assert.Equal(t, mt.Value, "0")
		default:
			assert.True(t, false)
		}
	}
}
