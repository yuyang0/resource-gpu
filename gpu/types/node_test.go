package types

import (
	"encoding/json"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/stretchr/testify/assert"
)

func TestNodeResource(t *testing.T) {
	nParams := map[string]any{
		"prod_count_map": nil,
	}
	n := &NodeResource{}
	err := n.Parse(nParams)
	assert.Nil(t, err)

	n = &NodeResource{}
	err = n.Parse(nil)
	assert.Nil(t, err)

	nParams = map[string]any{
		"prod_count_map1": nil,
	}
	n = &NodeResource{}
	err = n.Parse(nParams)
	assert.Nil(t, err)
	assert.Nil(t, n.ProdCountMap)

	nParams = map[string]any{
		"prod_count_map": ProdCountMap{
			"nvidia-3070": 4,
			"nvidia-3090": 2,
		},
	}
	n = &NodeResource{}
	err = n.Parse(nParams)
	assert.Nil(t, err)
	assert.Equal(t, n.ProdCountMap.TotalCount(), 6)
}

func TestNodeResourceRequest(t *testing.T) {
	req := &NodeResourceRequest{}
	err := req.Parse(nil)
	assert.Nil(t, err)
}

func TestJsonLoadNodeReqResp(t *testing.T) {
	j1 := `
{
	"prod_count_map": {
		"nvidia-3070": 4,
		":nvidia-3090": 2
	}
}
	`
	obj := resourcetypes.RawParams{}
	err := json.Unmarshal([]byte(j1), &obj)
	assert.Nil(t, err)
	req := &NodeResourceRequest{}
	err = req.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, req.Count(), 6)

	res := &NodeResource{}
	err = res.Parse(obj)
	assert.Nil(t, err)
	assert.Equal(t, res.Count(), 6)
}
