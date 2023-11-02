package bandwidth

import (
	"context"
	"fmt"
	"testing"

	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	cm := initBandwidth(context.Background(), t)
	assert.Equal(t, cm.name, cm.Name())
}

func initBandwidth(ctx context.Context, t *testing.T) *Plugin {
	config := coretypes.Config{
		Etcd: coretypes.EtcdConfig{
			Prefix: "/bandwidth",
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
	reqs := generateNodeResourceRequests(t, nums, index, "test", 100)
	info := &enginetypes.Info{NCPU: 8, MemTotal: 2048}
	names := []string{}
	for name, req := range reqs {
		_, err := cm.AddNode(ctx, name, req, info)
		assert.NoError(t, err)
		names = append(names, name)
	}
	t.Cleanup(func() {
		for name := range reqs {
			_, err := cm.RemoveNode(ctx, name)
			assert.NoError(t, err)
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
			_, err := cm.RemoveNode(ctx, name)
			assert.NoError(t, err)
		}
	})
	return names
}

func generateNodeResourceRequests(t *testing.T, nums int, index int, namePrefix string, bandwidth int64) map[string]plugintypes.NodeResourceRequest {
	infos := map[string]plugintypes.NodeResourceRequest{}
	for i := index; i < index+nums; i++ {
		info := plugintypes.NodeResourceRequest{
			"bandwidth": bandwidth,
		}
		infos[fmt.Sprintf("%s%v", namePrefix, i)] = info
	}
	return infos
}
