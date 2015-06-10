package command

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/heartbeatsjp/happo-agent/util"
)

func CmdAdd(c *cli.Context) {

	manage_request, err := util.BindManageParameter(c)
	data, err := json.Marshal(manage_request)
	if err != nil {
		log.Fatalf(err.Error())
	}

	resp, err := util.RequestToManageAPI(c.String("endpoint"), "/manage/add", data)
	if err != nil && resp == nil {
		log.Fatalf(err.Error())
	}
	if resp.StatusCode == http.StatusFound {
		log.Fatalf("Conflict!")
	} else if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed!")
	}
	log.Printf("Success.")
}
