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
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "gpu_used",
			"help":   "node used gpu.",
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
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.CapCount()),
			"key":    fmt.Sprintf("core.node.%s.gpu.capacity", safeNodename),
		},
		{
			"name":   "gpu_used",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.UsageCount()),
			"key":    fmt.Sprintf("core.node.%s.gpu.used", safeNodename),
		},
	}

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
