package node

import (
	"github.com/projecteru2/resource-gpu/cmd"
	"github.com/projecteru2/resource-gpu/gpu"

	"github.com/mitchellh/mapstructure"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func AddNode() *cli.Command {
	return &cli.Command{
		Name:   binary.AddNodeCommand,
		Usage:  "add node",
		Action: addNode,
	}
}

func RemoveNode() *cli.Command {
	return &cli.Command{
		Name:   binary.RemoveNodeCommand,
		Usage:  "remove node",
		Action: removeNode,
	}
}

func addNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		engineInfo := in.RawParams("info")
		resource := in.RawParams("resource")
		info := &enginetypes.Info{}
		if err := mapstructure.Decode(engineInfo, info); err != nil {
			return nil, err
		}
		return s.AddNode(c.Context, nodename, resource, info)
	})
}

func removeNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		return nil, s.RemoveNode(c.Context, nodename)
	})
}
