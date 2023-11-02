package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-bandwidth/bandwidth"
)

var (
	ConfigPath      string
	EmbeddedStorage bool
)

func Serve(c *cli.Context, f func(s *bandwidth.Plugin, in resourcetypes.RawParams) (interface{}, error)) error {
	config, err := utils.LoadConfig(ConfigPath)
	if err != nil {
		return cli.Exit(err, 128)
	}

	var t *testing.T
	if EmbeddedStorage {
		t = &testing.T{}
	}

	s, err := bandwidth.NewPlugin(c.Context, config, t)
	if err != nil {
		return cli.Exit(err, 128)
	}

	in := resourcetypes.RawParams{}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintf(os.Stderr, "Bandwidth: failed decode input json: %s\n", err)
		fmt.Fprintf(os.Stderr, "Bandwidth: input: %v\n", in)
		return cli.Exit(err, 128)
	}

	if r, err := f(s, in); err != nil {
		fmt.Fprintf(os.Stderr, "Bandwidth: failed call function: %s\n", err)
		fmt.Fprintf(os.Stderr, "Bandwidth: input: %v\n", in)
		return cli.Exit(err, 128)
	} else if o, err := json.Marshal(r); err != nil {
		fmt.Fprintf(os.Stderr, "Bandwidth: failed encode return object: %s\n", err)
		fmt.Fprintf(os.Stderr, "Bandwidth: input: %v\n", in)
		fmt.Fprintf(os.Stderr, "Bandwidth: output: %v\n", r)
		return cli.Exit(err, 128)
	} else { //nolint
		fmt.Print(string(o))
	}
	return nil
}
