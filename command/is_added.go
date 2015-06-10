package command

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/util"
)

// Check host database
func CmdIs_added(c *cli.Context) {

	manage_request, err := util.BindManageParameter(c)
	data, err := json.Marshal(manage_request)
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
