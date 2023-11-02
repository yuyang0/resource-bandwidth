package bandwidth

import (
	"context"
	"errors"
	"testing"

	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/yuyang0/resource-bandwidth/bandwidth/types"
)

func TestCalculateDeploy(t *testing.T) {
	ctx := context.Background()
	cm := initBandwidth(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// invalid opts
	req := plugintypes.WorkloadResourceRequest{
		"bandwidth": -10,
	}
	_, err := cm.CalculateDeploy(ctx, node, 100, req)
	assert.True(t, errors.Is(err, types.ErrInvalidBandwidth))

	// non-existent node
	req = plugintypes.WorkloadResourceRequest{
		"bandwidth": 10,
	}
	_, err = cm.CalculateDeploy(ctx, "xxx", 100, req)
	assert.True(t, errors.Is(err, coretypes.ErrNodeNotExists))

	parse := func(d *plugintypes.CalculateDeployResponse) (eps []*types.EngineParams, wrs []*types.WorkloadResource) {
		assert.NotNil(t, d.EnginesParams)
		assert.NotNil(t, d.WorkloadsResource)
		for _, epRaw := range d.EnginesParams {
			ep := &types.EngineParams{}
			err := ep.Parse(epRaw)
			assert.Nil(t, err)
			eps = append(eps, ep)
		}
		for _, wrRaw := range d.WorkloadsResource {
			wr := &types.WorkloadResource{}
			err := wr.Parse(wrRaw)
			assert.Nil(t, err)
			wrs = append(wrs, wr)
		}
		return
	}
	// normal cases
	// 1. empty request
	d, err := cm.CalculateDeploy(ctx, node, 4, nil)
	assert.Nil(t, err)
	assert.NotNil(t, d.EnginesParams)
	eParams, wResources := parse(d)
	assert.Len(t, eParams, 4)
	assert.Len(t, wResources, 4)
	for i := 0; i < 4; i++ {
		assert.Equal(t, eParams[i].Average, int64(0))
		assert.Equal(t, wResources[i].Bandwidth, int64(0))
	}
	// has enough resource
	d, err = cm.CalculateDeploy(ctx, node, 10, req)
	assert.Nil(t, err)
	eParams, wResources = parse(d)
	assert.Len(t, eParams, 10)

	// don't have enough resource
	d, err = cm.CalculateDeploy(ctx, node, 11, req)
	assert.Error(t, err)
}

func TestCalculateRealloc(t *testing.T) {
	ctx := context.Background()
	cm := initBandwidth(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]

	// set capacity
	resource := plugintypes.NodeResource{
		"bandwidth": 10,
	}

	_, err := cm.SetNodeResourceUsage(ctx, node, nil, resource, nil, false, true)
	assert.Nil(t, err)

	origin := plugintypes.WorkloadResource{}
	req := plugintypes.WorkloadResourceRequest{}

	// non-existent node
	_, err = cm.CalculateRealloc(ctx, "xxx", origin, req)
	assert.True(t, errors.Is(err, coretypes.ErrNodeNotExists))

	parse := func(d *plugintypes.CalculateReallocResponse) (*types.EngineParams, *types.WorkloadResource, *types.WorkloadResource) {
		assert.NotNil(t, d.EngineParams)
		assert.NotNil(t, d.WorkloadResource)
		ep := &types.EngineParams{}
		err := ep.Parse(d.EngineParams)
		assert.Nil(t, err)

		wr := &types.WorkloadResource{}
		err = wr.Parse(d.WorkloadResource)
		assert.Nil(t, err)

		dwr := &types.WorkloadResource{}
		err = dwr.Parse(d.DeltaResource)
		assert.Nil(t, err)
		return ep, wr, dwr
	}
	// normal cases
	// 1. empty request and resource
	d, err := cm.CalculateRealloc(ctx, node, nil, nil)
	assert.Nil(t, err)
	eParams, wResource, dResource := parse(d)
	assert.Equal(t, eParams.Bandwidth(), int64(0))
	assert.Equal(t, wResource.Bandwidth, int64(0))
	assert.Equal(t, dResource.Bandwidth, int64(0))
	// 2. empty request
	origin = plugintypes.WorkloadResource{
		"bandwidth": 10,
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, nil)
	assert.Nil(t, err)
	eParams, wResource, dResource = parse(d)

	assert.Equal(t, eParams.Average, int64(10))

	assert.Equal(t, wResource.Bandwidth, int64(10))

	assert.Equal(t, dResource.Bandwidth, int64(0))
	// 3. overwirte resource with request
	origin = plugintypes.WorkloadResource{
		"bandwidth": 10,
	}

	req = plugintypes.WorkloadResourceRequest{
		"bandwidth": 20,
	}
	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams, wResource, dResource = parse(d)
	assert.Equal(t, eParams.Average, int64(30))
	assert.Equal(t, wResource.Bandwidth, int64(30))
	assert.Equal(t, dResource.Bandwidth, int64(20))

	// decrease bandwidth
	origin = plugintypes.WorkloadResource{
		"bandwidth": 30,
	}
	req = plugintypes.WorkloadResourceRequest{
		"bandwidth": -20,
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams, wResource, dResource = parse(d)

	assert.Equal(t, eParams.Average, int64(10))
	assert.Equal(t, wResource.Bandwidth, int64(10))
	assert.Equal(t, dResource.Bandwidth, int64(-20))
	// smaller negative count
	origin = plugintypes.WorkloadResource{
		"bandwidth": 30,
	}
	req = plugintypes.WorkloadResourceRequest{
		"bandwidth": -40,
	}

	d, err = cm.CalculateRealloc(ctx, node, origin, req)
	assert.Nil(t, err)
	eParams, wResource, dResource = parse(d)

	assert.Equal(t, eParams.Average, int64(0))
	assert.Equal(t, wResource.Bandwidth, int64(0))
	assert.Equal(t, dResource.Bandwidth, int64(-30))
}

func TestCalculateRemap(t *testing.T) {
	ctx := context.Background()
	cm := initBandwidth(ctx, t)
	nodes := generateNodes(ctx, t, cm, 1, 0)
	node := nodes[0]
	d, err := cm.CalculateRemap(ctx, node, nil)

	assert.NoError(t, err)
	assert.Nil(t, d.EngineParamsMap)
}
