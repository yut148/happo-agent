package command

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/util"
)

// CmdIsAdded implements subcommand `is_added`. Check host database
func CmdIsAdded(c *cli.Context) {

	manageRequest, err := util.BindManageParameter(c)
	data, err := json.Marshal(manageRequest)
	if err != nil {
		log.Fatalf(err.Error())
	}

	resp, err := util.RequestToManageAPI(c.String("endpoint"), "/manage/is_added", data)
	if err != nil && resp == nil {
		log.Fatalf(err.Error())
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("Not found.")
		os.Exit(1)
	} else if resp.StatusCode == http.StatusFound {
		log.Printf("Found.")
		os.Exit(0)
	}
	log.Printf("Unknown Status")
	os.Exit(2)
}
