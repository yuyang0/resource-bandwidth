package bandwidth

import (
	"github.com/yuyang0/resource-bandwidth/bandwidth"
	"github.com/yuyang0/resource-bandwidth/cmd"

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
	return cmd.Serve(c, func(s *bandwidth.Plugin, _ resourcetypes.RawParams) (interface{}, error) {
		return s.Name(), nil
	})
}
