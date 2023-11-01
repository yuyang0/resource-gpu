package node

import (
	"encoding/json"

	"github.com/yuyang0/resource-gpu/cmd"
	"github.com/yuyang0/resource-gpu/gpu"

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
		eInfoBytes, err := json.Marshal(engineInfo)
		if err != nil {
			return nil, err
		}
		resource := in.RawParams("resource")
		info := &enginetypes.Info{}
		if err := json.Unmarshal(eInfoBytes, info); err != nil {
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
		return s.RemoveNode(c.Context, nodename)
	})
}
