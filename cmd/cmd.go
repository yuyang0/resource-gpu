package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-gpu/gpu"
)

var (
	ConfigPath      string
	EmbeddedStorage bool
)

func Serve(c *cli.Context, f func(s *gpu.Plugin, in resourcetypes.RawParams) (interface{}, error)) error {
	config, err := utils.LoadConfig(ConfigPath)
	if err != nil {
		return cli.Exit(err, 128)
	}

	var t *testing.T
	if EmbeddedStorage {
		t = &testing.T{}
	}

	s, err := gpu.NewPlugin(c.Context, config, t)
	if err != nil {
		return cli.Exit(err, 128)
	}

	in := resourcetypes.RawParams{}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintf(os.Stderr, "GPU: failed decode input json: %s\n", err)
		fmt.Fprintf(os.Stderr, "GPU: input: %v\n", in)
		return cli.Exit(err, 128)
	}

	if r, err := f(s, in); err != nil {
		fmt.Fprintf(os.Stderr, "GPU: failed call function: %s\n", err)
		fmt.Fprintf(os.Stderr, "GPU: input: %v\n", in)
		return cli.Exit(err, 128)
	} else if o, err := json.Marshal(r); err != nil {
		fmt.Fprintf(os.Stderr, "GPU: failed encode return object: %s\n", err)
		fmt.Fprintf(os.Stderr, "GPU: input: %v\n", in)
		fmt.Fprintf(os.Stderr, "GPU: output: %v\n", o)
		return cli.Exit(err, 128)
	} else {
		fmt.Print(string(o))
	}
	return nil
}
