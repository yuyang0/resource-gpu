package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/utils"
	"github.com/projecteru2/resource-gpu/gpu"
	zerolog "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
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
	if err := log.SetupLog(c.Context, config.LogLevel, config.SentryDSN); err != nil {
		zerolog.Fatal().Err(err).Send()
	}
	s, err := gpu.NewPlugin(c.Context, config, t)
	if err != nil {
		return cli.Exit(err, 128)
	}

	in := resourcetypes.RawParams{}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		return cli.Exit(err, 128)
	}

	if r, err := f(s, in); err != nil {
		return cli.Exit(err, 128)
	} else if o, err := json.Marshal(r); err != nil {
		return cli.Exit(err, 128)
	} else {
		fmt.Print(string(o))
	}
	return nil
}
