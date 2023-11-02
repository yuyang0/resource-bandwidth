package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// EngineParams .
type EngineParams struct {
	Average int64 `json:"average" mapstructure:"average"`
	Peak    int64 `json:"peak" mapstructure:"peak"`
}

func (ep *EngineParams) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"average": ep.Average,
		"peak":    ep.Peak,
	}
}

func (ep *EngineParams) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, ep)
}

func (ep *EngineParams) Count() int64 {
	return ep.Average
}

func (ep *EngineParams) DeepCopy() *EngineParams {
	return &EngineParams{
		Average: ep.Average,
		Peak:    ep.Peak,
	}
}

func (ep *EngineParams) Sub(ep1 *EngineParams) {
	ep.Average -= ep1.Average
	ep.Peak -= ep1.Peak
}

func (ep *EngineParams) Add(ep1 *EngineParams) {
	ep.Average += ep1.Average
	ep.Peak += ep1.Peak
}
