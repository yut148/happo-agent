package command

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
)

// CmdIsAdded implements subcommand `is_added`. Check host database
func CmdIsAdded(c *cli.Context) error {
	if c.String("endpoint") == halib.DefaultAPIEndpoint {
		return cli.NewExitError("ERROR: endpoint must set with args or environment variable", 1)
	}

	manageRequest, err := util.BindManageParameter(c)
	data, err := json.Marshal(manageRequest)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	resp, err := util.RequestToManageAPI(c.String("endpoint"), "/manage/is_added", data)
	if err != nil && resp == nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if resp.StatusCode == http.StatusNotFound {
		return cli.NewExitError("Not found.", 1)
	} else if resp.StatusCode == http.StatusFound {
		log.Printf("Found.")
		return nil
	}
	return cli.NewExitError("Unknown Status", 2)
}
