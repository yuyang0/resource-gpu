package node

import (
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-gpu/cmd"
	"github.com/projecteru2/resource-gpu/gpu"
	"github.com/urfave/cli/v2"
)

func GetMostIdleNode() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMostIdleNodeCommand,
		Usage:  "get most idle node",
		Action: getMostIdleNode,
	}
}

func getMostIdleNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodenames := in.StringSlice("nodenames")
		if len(nodenames) == 0 {
			return nil, types.ErrEmptyNodeName
		}

		return s.GetMostIdleNode(c.Context, nodenames)
	})
}
