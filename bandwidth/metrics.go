package bandwidth

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
			"name":   "bandwidth_capacity",
			"help":   "node available bandwidth.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "bandwidth_used",
			"help":   "node used bandwidth.",
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
	var metrics []map[string]any
	metrics = append(metrics, map[string]any{
		"name":   "bandwidth_capacity",
		"labels": []string{podname, nodename},
		"value":  fmt.Sprintf("%+v", nodeResourceInfo.Capacity.Bandwidth),
		"key":    fmt.Sprintf("core.node.%s.bandwidth.capacity", safeNodename),
	})
	metrics = append(metrics, map[string]any{
		"name":   "bandwidth_used",
		"labels": []string{podname, nodename},
		"value":  fmt.Sprintf("%+v", nodeResourceInfo.Usage.Bandwidth),
		"key":    fmt.Sprintf("core.node.%s.bandwidth.used", safeNodename),
	})

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
