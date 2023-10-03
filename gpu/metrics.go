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
			"name":   "gpu_capacity",
			"help":   "node available gpu.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename", "product"},
		},
		{
			"name":   "gpu_used",
			"help":   "node used gpu.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename", "product"},
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
	var metrics []map[string]any
	for prod, count := range nodeResourceInfo.Capacity.ProdCountMap {
		metrics = append(metrics, map[string]any{
			"name":   "gpu_capacity",
			"labels": []string{podname, nodename, prod},
			"value":  fmt.Sprintf("%+v", count),
			"key":    fmt.Sprintf("core.node.%s.gpu.capacity", safeNodename),
		})
		usageCount := nodeResourceInfo.Usage.ProdCountMap[prod]
		metrics = append(metrics, map[string]any{
			"name":   "gpu_used",
			"labels": []string{podname, nodename, prod},
			"value":  fmt.Sprintf("%+v", usageCount),
			"key":    fmt.Sprintf("core.node.%s.gpu.used", safeNodename),
		})
	}

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
