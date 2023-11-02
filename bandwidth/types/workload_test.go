package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	resourcetypes "github.com/projecteru2/core/resource/types"
)

func TestWorkloadResource(t *testing.T) {
	wr := &WorkloadResource{}
	err := wr.Parse(nil)
	assert.Nil(t, err)
}

func TestWorkloadResourceRequest(t *testing.T) {
	// empty request
	req := &WorkloadResourceRequest{}
	err := req.Parse(nil)
	assert.Nil(t, err)
	assert.Nil(t, req.Validate())

	params := resourcetypes.RawParams{
		"bandwidth": 100,
	}
	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Nil(t, err)
	assert.Equal(t, req.Bandwidth, int64(100))

	// invalid request
	params = resourcetypes.RawParams{
		"bandwidth": -100,
	}

	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Nil(t, err)
	assert.Error(t, req.Validate())
}

func TestJsonLoad(t *testing.T) {
	j1 := `
{
	"bandwidth": 100
}
	`
	obj := resourcetypes.RawParams{}
	err := json.Unmarshal([]byte(j1), &obj)
	assert.Nil(t, err)
	req := &WorkloadResourceRequest{}
	err = req.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, req.Bandwidth, int64(100))

	res := &WorkloadResource{}
	err = res.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, res.Bandwidth, int64(100))
}
