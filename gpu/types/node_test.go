package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeResource(t *testing.T) {
	nParams := map[string]any{
		"gpu_map": nil,
	}
	n := &NodeResource{}
	err := n.Parse(nParams)
	assert.Nil(t, err)

	n = &NodeResource{}
	err = n.Parse(nil)
	assert.Nil(t, err)

	nParams = map[string]any{
		"gpu_map1": nil,
	}
	n = &NodeResource{}
	err = n.Parse(nParams)
	assert.Nil(t, err)
	assert.Nil(t, n.GPUMap)

	nParams = map[string]any{
		"gpu_map": GPUMap{
			"add1": {
				Address: "addr1",
			},
		},
	}
}
