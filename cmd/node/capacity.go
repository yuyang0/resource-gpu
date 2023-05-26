package node

import (
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-gpu/cmd"
	"github.com/yuyang0/resource-gpu/gpu"
)

func GetNodesDeployCapacity() *cli.Command {
	return &cli.Command{
		Name:   binary.GetNodesDeployCapacityCommand,
		Usage:  "get deploy capacity",
		Action: getNodesDeployCapacity,
	}
}

func getNodesDeployCapacity(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodenames := in.StringSlice("nodenames")
		if len(nodenames) == 0 {
			return nil, types.ErrEmptyNodeName
		}

		workloadResource := in.RawParams("workload_resource")
		return s.GetNodesDeployCapacity(c.Context, nodenames, workloadResource)
	})
}

func SetNodeResourceCapacity() *cli.Command {
	return &cli.Command{
		Name:   binary.SetNodeResourceCapacityCommand,
		Usage:  "set node capacity",
		Action: setNodeResourceCapacity,
	}
}

func setNodeResourceCapacity(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		incr := in.Bool("incr")
		delta := in.Bool("delta")
		resource := in.RawParams("resource")
		resourceRequest := in.RawParams("resource_request")
		return s.SetNodeResourceCapacity(c.Context, nodename, resource, resourceRequest, delta, incr)
	})
}
