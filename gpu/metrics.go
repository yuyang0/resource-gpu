package gpu

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
)

// GetMetricsDescription .
func (p Plugin) GetMetricsDescription(context.Context) (*plugintypes.GetMetricsDescriptionResponse, error) {
	resp := &plugintypes.GetMetricsDescriptionResponse{}
	return resp, mapstructure.Decode([]map[string]interface{}{
		{
			"name":   "cpu_map",
			"help":   "node available cpu.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename", "cpuid"},
		},
		{
			"name":   "cpu_used",
			"help":   "node used cpu.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "memory_capacity",
			"help":   "node available memory.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "memory_used",
			"help":   "node used memory.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
	}, resp)
}

// GetMetrics .
func (p Plugin) GetMetrics(ctx context.Context, podname, nodename string) (*plugintypes.GetMetricsResponse, error) {
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}
	safeNodename := strings.ReplaceAll(nodename, ".", "_")
	metrics := []map[string]interface{}{
		{
			"name":   "gpu_capacity",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", len(nodeResourceInfo.Capacity.GPUMap)),
			"key":    fmt.Sprintf("core.node.%s.gpu", safeNodename),
		},
		{
			"name":   "gpu_used",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", len(nodeResourceInfo.Usage.GPUMap)),
			"key":    fmt.Sprintf("core.node.%s.gpu.used", safeNodename),
		},
	}

	for addr, info := range nodeResourceInfo.Usage.GPUMap {
		metrics = append(metrics, map[string]interface{}{
			"name":   "gpu_map",
			"labels": []string{podname, nodename, addr},
			"value":  fmt.Sprintf("%+v", 100),
			"key":    fmt.Sprintf("core.node.%s.gpu.%s(%s)", safeNodename, addr, info.Product),
		})
	}

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
