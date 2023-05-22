package calculate

import (
	"github.com/mitchellh/mapstructure"
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-gpu/cmd"
	"github.com/projecteru2/resource-gpu/gpu"
	"github.com/urfave/cli/v2"
)

func CalculateRemap() *cli.Command { //nolint
	return &cli.Command{
		Name:   binary.CalculateRemapCommand,
		Usage:  "remap resource",
		Action: calculateRemap,
	}
}

func calculateRemap(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadsResource := map[string]resourcetypes.RawParams{}
		for ID, data := range in.RawParams("workloads_resource") {
			workloadsResource[ID] = resourcetypes.RawParams{}
			_ = mapstructure.Decode(data, workloadsResource[ID])
		}
		// NO NEED REMAP VOLUME
		return s.CalculateRemap(c.Context, nodename, workloadsResource)
	})
}
