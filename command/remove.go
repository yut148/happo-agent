package command

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

// CmdRemove implements subcommand `remove`
func CmdRemove(c *cli.Context) error {
	if c.String("endpoint") == halib.DefaultAPIEndpoint {
		return cli.NewExitError("ERROR: endpoint must set with args or environment variable", 1)
	}

	manageRequest, err := util.BindManageParameter(c)
	data, err := json.Marshal(manageRequest)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	resp, err := util.RequestToManageAPI(c.String("endpoint"), "/manage/remove", data)
	if err != nil && resp == nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if resp.StatusCode != http.StatusOK {
		return cli.NewExitError("Failed!", 1)
	}
	fmt.Println("Success.")
	return nil
}
