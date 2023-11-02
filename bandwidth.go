package main

import (
	"context"
	"fmt"
	"os"

	"github.com/projecteru2/core/resource/plugins"
	coretypes "github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
	bdlib "github.com/yuyang0/resource-bandwidth/bandwidth"
	"github.com/yuyang0/resource-bandwidth/cmd"
	"github.com/yuyang0/resource-bandwidth/cmd/bandwidth"
	"github.com/yuyang0/resource-bandwidth/cmd/calculate"
	"github.com/yuyang0/resource-bandwidth/cmd/metrics"
	"github.com/yuyang0/resource-bandwidth/cmd/node"
	"github.com/yuyang0/resource-bandwidth/version"
)

func NewPlugin(ctx context.Context, config coretypes.Config) (plugins.Plugin, error) {
	p, err := bdlib.NewPlugin(ctx, config, nil)
	return p, err
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.String())
	}

	app := cli.NewApp()
	app.Name = version.NAME
	app.Usage = "Run eru resource Bandwidth plugin"
	app.Version = version.VERSION
	app.Commands = []*cli.Command{
		bandwidth.Name(),
		metrics.Description(),
		metrics.GetMetrics(),

		node.AddNode(),
		node.RemoveNode(),
		node.GetNodesDeployCapacity(),
		node.SetNodeResourceCapacity(),
		node.GetNodeResourceInfo(),
		node.SetNodeResourceInfo(),
		node.SetNodeResourceUsage(),
		node.GetMostIdleNode(),
		node.FixNodeResource(),

		calculate.CalculateDeploy(),
		calculate.CalculateRealloc(),
		calculate.CalculateRemap(),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Value:       "bandwidth.yaml",
			Usage:       "config file path for plugin, in yaml",
			Destination: &cmd.ConfigPath,
			EnvVars:     []string{"ERU_RESOURCE_CONFIG_PATH"},
		},
		&cli.BoolFlag{
			Name:        "embedded-storage",
			Usage:       "active embedded storage",
			Destination: &cmd.EmbeddedStorage,
		},
	}
	_ = app.Run(os.Args)
}
