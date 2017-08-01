package model

import (
	"log"
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/lib"
	"github.com/heartbeatsjp/happo-agent/util"
)

// --- Constant Values

// --- Method

func Inventory(inventory_request lib.InventoryRequest, r render.Render, params martini.Params) {
	var inventory_response lib.InventoryResponse

	if !util.Production {
		log.Printf("Inventory Command: %s %s\n", inventory_request.Command, inventory_request.CommandOption)
	}

	exitstatus, out, err := util.ExecCommandCombinedOutput(inventory_request.Command, inventory_request.CommandOption)
	if err != nil {
		r.JSON(http.StatusExpectationFailed, inventory_response)
		return
	}

	if exitstatus != 0 {
		inventory_response.ReturnCode = exitstatus
		inventory_response.ReturnValue = out
		r.JSON(http.StatusBadRequest, inventory_response)
		return
	}
	inventory_response.ReturnCode = exitstatus
	inventory_response.ReturnValue = out

	r.JSON(http.StatusOK, inventory_response)
}
