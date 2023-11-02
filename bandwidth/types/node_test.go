package types

import (
	"encoding/json"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/stretchr/testify/assert"
)

func TestNodeResource(t *testing.T) {
	nParams := map[string]any{}
	n := &NodeResource{}
	err := n.Parse(nParams)
	assert.Nil(t, err)

	n = &NodeResource{}
	err = n.Parse(nil)
	assert.Nil(t, err)

	nParams = map[string]any{
		"bandwidth1": 100,
	}
	n = &NodeResource{}
	err = n.Parse(nParams)
	assert.Nil(t, err)
	assert.Zero(t, n.Bandwidth)

	nParams = map[string]any{
		"bandwidth": 100,
	}
	n = &NodeResource{}
	err = n.Parse(nParams)
	assert.Nil(t, err)
	assert.Equal(t, n.Bandwidth, int64(100))
}

func TestNodeResourceRequest(t *testing.T) {
	req := &NodeResourceRequest{}
	err := req.Parse(nil)
	assert.Nil(t, err)
}

func TestJsonLoadNodeReqResp(t *testing.T) {
	j1 := `
{
	"bandwidth": 100
}
	`
	obj := resourcetypes.RawParams{}
	err := json.Unmarshal([]byte(j1), &obj)
	assert.Nil(t, err)
	req := &NodeResourceRequest{}
	err = req.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, req.Bandwidth, int64(100))

	res := &NodeResource{}
	err = res.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, res.Bandwidth, int64(100))
}
