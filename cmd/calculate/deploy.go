package calculate

import (
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-gpu/cmd"
	"github.com/yuyang0/resource-gpu/gpu"
)

func CalculateDeploy() *cli.Command { //nolint
	return &cli.Command{
		Name:   binary.CalculateDeployCommand,
		Usage:  "calculate deploy plan",
		Action: calculateDeploy,
	}
}

func calculateDeploy(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		deployCount := in.Int("deploy_count")

		workloadResourceRequest := in.RawParams("workload_resource_request")
		return s.CalculateDeploy(c.Context, nodename, deployCount, workloadResourceRequest)
	})
}
