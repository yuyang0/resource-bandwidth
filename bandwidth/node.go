package bandwidth

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/cockroachdb/errors"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"

	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	"github.com/sanity-io/litter"
	bdtypes "github.com/yuyang0/resource-bandwidth/bandwidth/types"
)

const (
	maxCapacity = 1000000
)

// AddNode .
func (p Plugin) AddNode(
	ctx context.Context, nodename string,
	resource plugintypes.NodeResourceRequest,
	info *enginetypes.Info,
) (
	*plugintypes.AddNodeResponse, error,
) {
	// try to get the node resource
	var err error
	if _, err = p.doGetNodeResourceInfo(ctx, nodename); err == nil {
		return nil, coretypes.ErrNodeExists
	}

	if !errors.IsAny(err, coretypes.ErrInvaildCount, coretypes.ErrNodeNotExists) {
		log.WithFunc("resource.bandwidth.AddNode").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	req := &bdtypes.NodeResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	capacity := bdtypes.NewNodeResource(req.Bandwidth)
	// try to fetch resource from info
	if info != nil && info.Resources != nil { //nolint
		if capacity.Bandwidth == 0 {
			if b, ok := info.Resources[p.name]; ok {
				err := json.Unmarshal(b, capacity)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	nodeResourceInfo := &bdtypes.NodeResourceInfo{
		Capacity: capacity,
		Usage:    bdtypes.NewNodeResource(0),
	}

	if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		return nil, err
	}
	return &plugintypes.AddNodeResponse{
		Capacity: nodeResourceInfo.Capacity.AsRawParams(),
		Usage:    nodeResourceInfo.Usage.AsRawParams(),
	}, nil
}

// RemoveNode .
func (p Plugin) RemoveNode(ctx context.Context, nodename string) (*plugintypes.RemoveNodeResponse, error) {
	var err error
	if _, err = p.store.Delete(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename)); err != nil {
		log.WithFunc("resource.bandwidth.RemoveNode").WithField("node", nodename).Error(ctx, err, "faield to delete node")
	}
	return &plugintypes.RemoveNodeResponse{}, err
}

// GetNodesDeployCapacity returns available nodes and total capacity
func (p Plugin) GetNodesDeployCapacity(
	ctx context.Context, nodenames []string,
	resource plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.GetNodesDeployCapacityResponse, error,
) {
	logger := log.WithFunc("resource.bandwidth.GetNodesDeployCapacity")
	req := &bdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodesDeployCapacityMap := map[string]*plugintypes.NodeDeployCapacity{}
	total := 0

	nodesResourceInfos, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfos {
		nodeDeployCapacity := p.doGetNodeDeployCapacity(nodeResourceInfo, req)
		if nodeDeployCapacity.Capacity > 0 {
			nodesDeployCapacityMap[nodename] = nodeDeployCapacity
			if total == math.MaxInt || nodeDeployCapacity.Capacity == math.MaxInt {
				total = math.MaxInt
			} else {
				total += nodeDeployCapacity.Capacity
			}
		}
	}
	return &plugintypes.GetNodesDeployCapacityResponse{
		NodeDeployCapacityMap: nodesDeployCapacityMap,
		Total:                 total,
	}, nil
}

// SetNodeResourceCapacity sets the amount of total resource info
func (p Plugin) SetNodeResourceCapacity(
	ctx context.Context, nodename string,
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	delta bool, incr bool,
) (
	*plugintypes.SetNodeResourceCapacityResponse, error,
) {
	logger := log.WithFunc("resource.bandwidth.SetNodeResourceCapacity").WithField("node", "nodename")
	req, nodeResource, _, err := p.parseNodeResourceInfos(resourceRequest, resource, nil)
	if err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}

	origin := nodeResourceInfo.Capacity
	before := origin.DeepCopy()

	if !delta && req != nil {
		req.LoadFromOrigin(origin, resourceRequest)
	}
	nodeResourceInfo.Capacity = p.calculateNodeResource(req, nodeResource, origin, nil, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return &plugintypes.SetNodeResourceCapacityResponse{
		Before: before.AsRawParams(),
		After:  nodeResourceInfo.Capacity.AsRawParams(),
	}, nil
}

// GetNodeResourceInfo .
func (p Plugin) GetNodeResourceInfo(
	ctx context.Context, nodename string,
	workloadsResource []plugintypes.WorkloadResource,
) (
	*plugintypes.GetNodeResourceInfoResponse, error,
) {
	nodeResourceInfo, _, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	return &plugintypes.GetNodeResourceInfoResponse{
		Capacity: nodeResourceInfo.Capacity.AsRawParams(),
		Usage:    nodeResourceInfo.Usage.AsRawParams(),
		Diffs:    diffs,
	}, nil
}

// SetNodeResourceInfo .
func (p Plugin) SetNodeResourceInfo(
	ctx context.Context, nodename string,
	capacity plugintypes.NodeResource,
	usage plugintypes.NodeResource,
) (
	*plugintypes.SetNodeResourceInfoResponse, error,
) {
	capacityResource := &bdtypes.NodeResource{}
	usageResource := &bdtypes.NodeResource{}
	if err := capacityResource.Parse(capacity); err != nil {
		return nil, err
	}
	if err := usageResource.Parse(usage); err != nil {
		return nil, err
	}
	resourceInfo := &bdtypes.NodeResourceInfo{
		Capacity: capacityResource,
		Usage:    usageResource,
	}

	return &plugintypes.SetNodeResourceInfoResponse{}, p.doSetNodeResourceInfo(ctx, nodename, resourceInfo)
}

// SetNodeResourceUsage .
func (p Plugin) SetNodeResourceUsage(
	ctx context.Context, nodename string,
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	workloadsResource []plugintypes.WorkloadResource,
	delta bool, incr bool,
) (
	*plugintypes.SetNodeResourceUsageResponse, error,
) {

	logger := log.WithFunc("resource.bandwidth.SetNodeResourceUsage").WithField("node", "nodename")
	req, nodeResource, wrksResource, err := p.parseNodeResourceInfos(resourceRequest, resource, workloadsResource)
	if err != nil {
		return nil, err
	}
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}

	origin := nodeResourceInfo.Usage
	before := origin.DeepCopy()

	nodeResourceInfo.Usage = p.calculateNodeResource(req, nodeResource, origin, wrksResource, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return &plugintypes.SetNodeResourceUsageResponse{
		Before: before.AsRawParams(),
		After:  nodeResourceInfo.Usage.AsRawParams(),
	}, nil
}

// GetMostIdleNode .
func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (*plugintypes.GetMostIdleNodeResponse, error) {
	var mostIdleNode string
	var minIdle = math.MaxFloat64

	nodesResourceInfo, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfo {
		var idle float64
		if nodeResourceInfo.CapBandwidth() > 0 {
			idle = float64(nodeResourceInfo.UsageBandwidth()) / float64(nodeResourceInfo.CapBandwidth())
		}

		if idle < minIdle {
			mostIdleNode = nodename
			minIdle = idle
		}
	}
	return &plugintypes.GetMostIdleNodeResponse{
		Nodename: mostIdleNode,
		Priority: priority,
	}, nil
}

// FixNodeResource .
// use workloadsReource to construct a new NodeResource, then use this NodeResource to repace Usage
func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*plugintypes.GetNodeResourceInfoResponse, error) {
	nodeResourceInfo, actuallyWorkloadsUsage, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	if len(diffs) != 0 {
		nodeResourceInfo.Usage = &bdtypes.NodeResource{
			Bandwidth: actuallyWorkloadsUsage.Bandwidth,
		}
		if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
			log.WithFunc("resource.bandwidth.FixNodeResource").Error(ctx, err)
			diffs = append(diffs, err.Error())
		}
	}
	return &plugintypes.GetNodeResourceInfoResponse{
		Capacity: nodeResourceInfo.Capacity.AsRawParams(),
		Usage:    nodeResourceInfo.Usage.AsRawParams(),
		Diffs:    diffs,
	}, nil
}

func (p Plugin) getNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*bdtypes.NodeResourceInfo, *bdtypes.WorkloadResource, []string, error) {
	logger := log.WithFunc("resource.bandwidth.getNodeResourceInfo").WithField("node", nodename)
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.Error(ctx, err)
		return nodeResourceInfo, nil, nil, err
	}

	actuallyWorkloadsUsage := &bdtypes.WorkloadResource{}
	for _, workloadResource := range workloadsResource {
		workloadUsage := &bdtypes.WorkloadResource{}
		if err := workloadUsage.Parse(workloadResource); err != nil {
			logger.Error(ctx, err)
			return nil, nil, nil, err
		}
		actuallyWorkloadsUsage.Add(workloadUsage)
	}

	diffs := []string{}

	if actuallyWorkloadsUsage.Bandwidth != nodeResourceInfo.UsageBandwidth() {
		diffs = append(diffs, fmt.Sprintf("node.BandwidthUsed != sum(workload.BandwidthRequest): %.2d != %.2d", nodeResourceInfo.UsageBandwidth(), actuallyWorkloadsUsage.Bandwidth))
	}

	return nodeResourceInfo, actuallyWorkloadsUsage, diffs, nil
}

func (p Plugin) doGetNodeResourceInfo(ctx context.Context, nodename string) (*bdtypes.NodeResourceInfo, error) {
	key := fmt.Sprintf(nodeResourceInfoKey, nodename)
	resp, err := p.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	r := &bdtypes.NodeResourceInfo{}
	switch resp.Count {
	case 0:
		return r, errors.Wrapf(coretypes.ErrNodeNotExists, "key: %s", nodename)
	case 1:
		if err := json.Unmarshal(resp.Kvs[0].Value, r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.Wrapf(coretypes.ErrInvaildCount, "key: %s", nodename)
	}
}

func (p Plugin) doGetNodesResourceInfo(ctx context.Context, nodenames []string) (map[string]*bdtypes.NodeResourceInfo, error) {
	keys := []string{}
	for _, nodename := range nodenames {
		keys = append(keys, fmt.Sprintf(nodeResourceInfoKey, nodename))
	}
	resps, err := p.store.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := map[string]*bdtypes.NodeResourceInfo{}

	for _, resp := range resps {
		r := &bdtypes.NodeResourceInfo{}
		if err := json.Unmarshal(resp.Value, r); err != nil {
			return nil, err
		}
		result[utils.Tail(string(resp.Key))] = r
	}
	return result, nil
}

func (p Plugin) doSetNodeResourceInfo(ctx context.Context, nodename string, resourceInfo *bdtypes.NodeResourceInfo) error {
	if err := resourceInfo.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(resourceInfo)
	if err != nil {
		return err
	}

	_, err = p.store.Put(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename), string(data))
	return err
}

func (p Plugin) doGetNodeDeployCapacity(nodeResourceInfo *bdtypes.NodeResourceInfo, req *bdtypes.WorkloadResourceRequest) *plugintypes.NodeDeployCapacity {
	availableResource := nodeResourceInfo.GetAvailableResource()

	capacityInfo := &plugintypes.NodeDeployCapacity{
		Weight:   1, // TODO why 1?
		Capacity: maxCapacity,
	}
	if req.Bandwidth == 0 { //nolint
		// if count equals to 0, then assign a big value to capacity
		capacityInfo.Capacity = maxCapacity
	} else {
		capacityInfo.Capacity = int(availableResource.Bandwidth / req.Bandwidth)
	}
	if nodeResourceInfo.CapBandwidth() > 0 {
		capacityInfo.Usage = float64(nodeResourceInfo.UsageBandwidth()) / float64(nodeResourceInfo.CapBandwidth())
		capacityInfo.Rate = float64(req.Bandwidth) / float64(nodeResourceInfo.CapBandwidth())
	}
	return capacityInfo
}

// 丢弃origin，完全用新数据重写
func (p Plugin) overwriteNodeResource(req *bdtypes.NodeResourceRequest, nodeResource *bdtypes.NodeResource, workloadsResource []*bdtypes.WorkloadResource) *bdtypes.NodeResource {
	resp := (&bdtypes.NodeResource{}).DeepCopy() // init nil pointer!
	if req != nil {
		nodeResource = &bdtypes.NodeResource{
			Bandwidth: req.Bandwidth,
		}
	}

	if nodeResource != nil {
		resp.Add(nodeResource)
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &bdtypes.NodeResource{
			Bandwidth: workloadResource.Bandwidth,
		}
		resp.Add(nodeResource)
	}
	return resp
}

// 增量更新
func (p Plugin) incrUpdateNodeResource(req *bdtypes.NodeResourceRequest, nodeResource *bdtypes.NodeResource, origin *bdtypes.NodeResource, workloadsResource []*bdtypes.WorkloadResource, incr bool) *bdtypes.NodeResource {
	resp := origin.DeepCopy()
	if req != nil {
		nodeResource = &bdtypes.NodeResource{
			Bandwidth: req.Bandwidth,
		}
	}

	if nodeResource != nil {
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &bdtypes.NodeResource{
			Bandwidth: workloadResource.Bandwidth,
		}
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
	}
	return resp
}

// calculateNodeResource priority: node resource request > node resource > workload resource args list
func (p Plugin) calculateNodeResource(req *bdtypes.NodeResourceRequest, nodeResource *bdtypes.NodeResource, origin *bdtypes.NodeResource, workloadsResource []*bdtypes.WorkloadResource, delta bool, incr bool) *bdtypes.NodeResource {
	// req, nodeResource, workloadResource只有一个会生效, 优先级是req, nodeResource, workloadsReource
	// 如果delta为false那就不考虑origin
	// 如果delta为true那就把3者中生效的那个加到origin上
	if origin == nil || !delta { // 重写
		return p.overwriteNodeResource(req, nodeResource, workloadsResource)
	} else { //nolint
		return p.incrUpdateNodeResource(req, nodeResource, origin, workloadsResource, incr)
	}
}

func (p Plugin) parseNodeResourceInfos(
	resourceRequest plugintypes.NodeResourceRequest,
	resource plugintypes.NodeResource,
	workloadsResource []plugintypes.WorkloadResource,
) (
	*bdtypes.NodeResourceRequest,
	*bdtypes.NodeResource,
	[]*bdtypes.WorkloadResource,
	error,
) {
	var req *bdtypes.NodeResourceRequest
	var nodeResource *bdtypes.NodeResource
	wrksResource := []*bdtypes.WorkloadResource{}

	if resourceRequest != nil {
		req = &bdtypes.NodeResourceRequest{}
		if err := req.Parse(resourceRequest); err != nil {
			return nil, nil, nil, err
		}
	}

	if resource != nil {
		nodeResource = &bdtypes.NodeResource{}
		if err := nodeResource.Parse(resource); err != nil {
			return nil, nil, nil, err
		}
	}

	for _, workloadResource := range workloadsResource {
		wrkResource := &bdtypes.WorkloadResource{}
		if err := wrkResource.Parse(workloadResource); err != nil {
			return nil, nil, nil, err
		}
		wrksResource = append(wrksResource, wrkResource)
	}

	return req, nodeResource, wrksResource, nil
}
