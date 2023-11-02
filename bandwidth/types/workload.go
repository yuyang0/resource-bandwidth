package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate Bandwidth workload resource
type WorkloadResource struct {
	Bandwidth int64 `json:"bandwidth" mapstructure:"bandwidth"`
}

func (w *WorkloadResource) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"bandwidth": w.Bandwidth,
	}
}
func (w *WorkloadResource) Validate() error {
	if w.Bandwidth < 0 {
		return ErrInvalidBandwidth
	}
	return nil
}

// ParseFromRawParams .
func (w *WorkloadResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, w)
}

// DeepCopy .
func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	res := &WorkloadResource{
		Bandwidth: w.Bandwidth,
	}
	return res
}

// Add .
func (w *WorkloadResource) Add(w1 *WorkloadResource) {
	w.Bandwidth += w1.Bandwidth
}

// Sub .
func (w *WorkloadResource) Sub(w1 *WorkloadResource) {
	w.Bandwidth -= w1.Bandwidth
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	Bandwidth int64 `json:"bandwidth" mapstructure:"bandwidth"`
}

// Validate .
func (w *WorkloadResourceRequest) Validate() error {
	if w.Bandwidth < 0 {
		return ErrInvalidBandwidth
	}
	return nil
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	return mapstructure.Decode(rawParams, w)
}

func (w *WorkloadResourceRequest) MergeFromResource(r *WorkloadResource) {
	w.Bandwidth += r.Bandwidth
	if w.Bandwidth < 0 {
		w.Bandwidth = 0
	}
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	return &WorkloadResourceRequest{
		Bandwidth: w.Bandwidth,
	}
}
