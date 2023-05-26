package gpu

import (
	"github.com/yuyang0/resource-gpu/cmd"
	"github.com/yuyang0/resource-gpu/gpu"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v2"
)

func Name() *cli.Command {
	return &cli.Command{
		Name:   "name",
		Usage:  "show name",
		Action: name,
	}
}

func name(c *cli.Context) error {
	return cmd.Serve(c, func(s *gpu.Plugin, _ resourcetypes.RawParams) (interface{}, error) {
		return s.Name(), nil
	})
}
