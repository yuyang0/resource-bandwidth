package bandwidth

import (
	"context"

	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	bdtypes "github.com/yuyang0/resource-bandwidth/bandwidth/types"
)

// CalculateDeploy .
func (p Plugin) CalculateDeploy(
	ctx context.Context, nodename string, deployCount int,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateDeployResponse, error,
) {
	logger := log.WithFunc("resource.bandwidth.CalculateDeploy").WithField("node", nodename)
	req := &bdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.WithField("node", nodename).Error(ctx, err)
		return nil, err
	}

	var enginesParams []*bdtypes.EngineParams
	var workloadsResource []*bdtypes.WorkloadResource

	enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, deployCount, req)
	if err != nil {
		return nil, err
	}

	epRaws := make([]resourcetypes.RawParams, 0, len(enginesParams))
	for _, ep := range enginesParams {
		epRaws = append(epRaws, ep.AsRawParams())
	}
	wrRaws := make([]resourcetypes.RawParams, 0, len(workloadsResource))
	for _, wr := range workloadsResource {
		wrRaws = append(wrRaws, wr.AsRawParams())
	}
	return &plugintypes.CalculateDeployResponse{
		EnginesParams:     epRaws,
		WorkloadsResource: wrRaws,
	}, nil
}

// CalculateRealloc .
func (p Plugin) CalculateRealloc(
	ctx context.Context, nodename string,
	resource plugintypes.WorkloadResource,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateReallocResponse, error,
) {
	req := &bdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	// realloc needs negative count, so don't need to validate

	originResource := &bdtypes.WorkloadResource{}
	if err := originResource.Parse(resource); err != nil {
		return nil, err
	}
	if err := originResource.Validate(); err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		log.WithFunc("resource.bandwidth.CalculateRealloc").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	// put resources back into the resource pool
	nodeResourceInfo.Usage.Sub(&bdtypes.NodeResource{
		Bandwidth: originResource.Bandwidth,
	})

	newReq := req.DeepCopy()
	newReq.MergeFromResource(originResource)

	if err = newReq.Validate(); err != nil {
		return nil, err
	}

	var enginesParams []*bdtypes.EngineParams
	var workloadsResource []*bdtypes.WorkloadResource
	if enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, 1, newReq); err != nil {
		return nil, err
	}

	engineParams := enginesParams[0]
	newResource := workloadsResource[0]

	deltaWorkloadResource := newResource.DeepCopy()
	deltaWorkloadResource.Sub(originResource)

	return &plugintypes.CalculateReallocResponse{
		EngineParams:     engineParams.AsRawParams(),
		DeltaResource:    deltaWorkloadResource.AsRawParams(),
		WorkloadResource: newResource.AsRawParams(),
	}, nil
}

// CalculateRemap .
func (p Plugin) CalculateRemap(context.Context, string, map[string]plugintypes.WorkloadResource) (*plugintypes.CalculateRemapResponse, error) {
	return &plugintypes.CalculateRemapResponse{
		EngineParamsMap: nil,
	}, nil
}

func (p Plugin) doAlloc(_ *bdtypes.NodeResourceInfo, deployCount int, req *bdtypes.WorkloadResourceRequest) ([]*bdtypes.EngineParams, []*bdtypes.WorkloadResource, error) { //nolint
	enginesParams := []*bdtypes.EngineParams{}
	workloadsResource := []*bdtypes.WorkloadResource{}

	for i := 0; i < deployCount; i++ {
		workloadsResource = append(workloadsResource, &bdtypes.WorkloadResource{
			Bandwidth: req.Bandwidth,
		})
		enginesParams = append(enginesParams, &bdtypes.EngineParams{
			Average: req.Bandwidth,
			Peak:    req.Bandwidth * 2,
		})
	}
	return enginesParams, workloadsResource, nil
}
