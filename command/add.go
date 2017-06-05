package command

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/util"
)

func CmdAdd(c *cli.Context) error {

	manage_request, err := util.BindManageParameter(c)
	data, err := json.Marshal(manage_request)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	resp, err := util.RequestToManageAPI(c.String("endpoint"), "/manage/add", data)
	if err != nil && resp == nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if resp.StatusCode == http.StatusFound {
		return cli.NewExitError("Conflict!", 1)
	} else if resp.StatusCode != http.StatusOK {
		return cli.NewExitError("Failed!", 1)
	}
	log.Printf("Success.")
	return nil
}
