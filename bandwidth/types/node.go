package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// NodeResource indicate node cpumem resource
type NodeResource struct {
	Bandwidth int64 `json:"bandwidth" mapstructure:"bandwidth"`
}

func NewNodeResource(bd int64) *NodeResource {
	return &NodeResource{
		Bandwidth: bd,
	}
}

func (r *NodeResource) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"bandwidth": r.Bandwidth,
	}
}

// Parse .
func (r *NodeResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, r)
}

func (r *NodeResource) Validate() error {
	if r.Bandwidth < 0 {
		return ErrInvalidBandwidth
	}
	return nil
}

// DeepCopy .
func (r *NodeResource) DeepCopy() *NodeResource {
	res := &NodeResource{
		Bandwidth: r.Bandwidth,
	}
	return res
}

// Add .
func (r *NodeResource) Add(r1 *NodeResource) {
	r.Bandwidth += r1.Bandwidth
}

// Sub .
func (r *NodeResource) Sub(r1 *NodeResource) {
	r.Bandwidth -= r1.Bandwidth
}

// NodeResourceInfo indicate cpumem capacity and usage
type NodeResourceInfo struct {
	Capacity *NodeResource `json:"capacity"`
	Usage    *NodeResource `json:"usage"`
}

func (n *NodeResourceInfo) CapCount() int64 {
	return n.Capacity.Bandwidth
}

func (n *NodeResourceInfo) UsageCount() int64 {
	return n.Usage.Bandwidth
}

// DeepCopy .
func (n *NodeResourceInfo) DeepCopy() *NodeResourceInfo {
	return &NodeResourceInfo{
		Capacity: n.Capacity.DeepCopy(),
		Usage:    n.Usage.DeepCopy(),
	}
}

func (n *NodeResourceInfo) Validate() error {
	if err := n.Capacity.Validate(); err != nil {
		return err
	}
	return n.Usage.Validate()
}

func (n *NodeResourceInfo) GetAvailableResource() *NodeResource {
	availableResource := n.Capacity.DeepCopy()
	availableResource.Sub(n.Usage)

	return availableResource
}

// NodeResourceRequest includes all possible fields passed by eru-core for editing node, it not parsed!
type NodeResourceRequest struct {
	Bandwidth int64 `json:"bandwidth" mapstructure:"bandwidth"`
}

func (n *NodeResourceRequest) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, n)
}

func (n *NodeResourceRequest) Validate() error {
	if n.Bandwidth < 0 {
		return ErrInvalidBandwidth
	}
	return nil
}

func (n *NodeResourceRequest) Count() int64 {
	return n.Bandwidth
}

// Merge fields to NodeResourceRequest.
func (n *NodeResourceRequest) LoadFromOrigin(nodeResource *NodeResource, resourceRequest resourcetypes.RawParams) {
	if n == nil {
		return
	}
	if !resourceRequest.IsSet("bandwidth") {
		n.Bandwidth = nodeResource.Bandwidth
	}
}
