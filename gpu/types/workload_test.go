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

	params := resourcetypes.RawParams{
		"prod_count_map": ProdCountMap{
			"nvidia-3070": 4,
			"nvidia-3090": 2,
		},
	}
	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Nil(t, err)
	assert.Equal(t, req.Count(), 6)

	// invalid request
	params = resourcetypes.RawParams{
		"prod_count_map": ProdCountMap{
			"nvidia-3070": 4,
			"nvidia-3090": 2,
			"  ":          1,
		},
	}

	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Error(t, req.Validate())

	params = resourcetypes.RawParams{
		"prod_count_map": ProdCountMap{
			"nvidia-3070": 4,
			"nvidia-3090": -1,
		},
	}

	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Error(t, req.Validate())
}
