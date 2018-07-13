package command

import (
	"fmt"

	"io/ioutil"

	"encoding/json"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

//CmdResolveAlias implements subcommand `resolve_alias`
func CmdResolveAlias(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return cli.NewExitError("missing argument", 1)
	}

	alias := c.Args().First()
	bastionEndpoint := c.String("bastion-endpoint")

	res, err := util.RequestToAutoScalingResolveAPI(bastionEndpoint, alias)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	var autoScalingResolveResponce halib.AutoScalingResolveResponse
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if err := json.Unmarshal(data, &autoScalingResolveResponce); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	fmt.Println(autoScalingResolveResponce.IP)
	return nil
}
