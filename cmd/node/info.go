package node

import (
	"github.com/cockroachdb/errors"
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-gpu/cmd"
	"github.com/yuyang0/resource-gpu/gpu"
)

func GetNodeResourceInfo() *cli.Command {
	return &cli.Command{
		Name:   binary.GetNodeResourceInfoCommand,
		Usage:  "get node resource info",
		Action: getNodeResourceInfo,
	}
}

func getNodeResourceInfo(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadsResource := in.SliceRawParams("workloads_resource")
		r, err := s.GetNodeResourceInfo(c.Context, nodename, workloadsResource)
		// when ETCD key doesn't exist, then return an empty NodeResourceInfo value
		if err == nil || errors.Is(err, coretypes.ErrNodeNotExists) {
			return r, nil
		}
		return r, err
	})
}

func SetNodeResourceInfo() *cli.Command {
	return &cli.Command{
		Name:   binary.SetNodeResourceInfoCommand,
		Usage:  "set node resource info",
		Action: setNodeResourceInfo,
	}
}

func setNodeResourceInfo(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		capacity := in.RawParams("capacity")
		usage := in.RawParams("usage")

		return nil, s.SetNodeResourceInfo(c.Context, nodename, capacity, usage)
	})
}

func FixNodeResource() *cli.Command {
	return &cli.Command{
		Name:   binary.FixNodeResourceCommand,
		Usage:  "fix node resource",
		Action: fixNodeResource,
	}
}

func fixNodeResource(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadsResource := in.SliceRawParams("workloads_resource")
		return s.FixNodeResource(c.Context, nodename, workloadsResource)
	})
}
