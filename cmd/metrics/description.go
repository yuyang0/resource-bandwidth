package metrics

import (
	"github.com/yuyang0/resource-bandwidth/bandwidth"
	"github.com/yuyang0/resource-bandwidth/cmd"

	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v2"
)

func Description() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMetricsDescriptionCommand,
		Usage:  "show metrics descriptions",
		Action: description,
	}
}

func description(c *cli.Context) error {
	return cmd.Serve(c, func(s *bandwidth.Plugin, _ resourcetypes.RawParams) (interface{}, error) {
		return s.GetMetricsDescription(c.Context)
	})
}
