package bandwidth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsDescription(t *testing.T) {
	ctx := context.Background()
	cm := initBandwidth(ctx, t)
	md, err := cm.GetMetricsDescription(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, md)
	assert.Len(t, *md, 2)
}

func TestGetMetrics(t *testing.T) {
	ctx := context.Background()
	cm := initBandwidth(ctx, t)
	_, err := cm.GetMetrics(ctx, "", "")
	assert.Error(t, err)

	nodes := generateNodes(ctx, t, cm, 1, -1)
	resp, err := cm.GetMetrics(ctx, "testpod", nodes[0])
	assert.NoError(t, err)
	for _, mt := range *resp {
		assert.Len(t, mt.Labels, 2)
		assert.Equal(t, mt.Labels[0], "testpod")
		assert.Equal(t, mt.Labels[1], nodes[0])
		switch mt.Name {
		case "bandwidth_capacity":
			assert.Equal(t, mt.Value, "100")
		case "bandwidth_used":
			assert.Equal(t, mt.Value, "0")
		default:
			assert.True(t, false)
		}
	}
}
