package types

import (
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

	// invalid request
	params := resourcetypes.RawParams{
		"count": 3,
		"gpus": []GPUInfo{
			{
				Product: "3070",
			},
			{
				Product: "3090",
			},
		},
	}
	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Error(t, req.Validate())
}
