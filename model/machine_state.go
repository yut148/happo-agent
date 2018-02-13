package model

import (
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/util"
)

// ListMachieState returns saved machine states
func ListMachieState(r render.Render) {
	log := util.HappoAgentLogger()

	var keys []string
	err := db.DB.View(func(tx *bolt.Tx) error {
		bucket := db.MachineStateBucket(tx)
		bucket.ForEach(func(key, value []byte) error {
			keys = append(keys, string(key))
			return nil
		})
		return nil
	})
	if err != nil {
		log.Error(err)
		r.JSON(http.StatusNoContent, map[string]string{"error": err.Error()})
		return
	}

	if len(keys) == 0 {
		r.JSON(http.StatusNotFound, map[string][]string{"keys": []string{}})
		return
	}
	r.JSON(http.StatusOK, map[string][]string{"keys": keys})
}

// GetMachineState returns saved specified machine state
func GetMachineState(r render.Render, params martini.Params) {
	log := util.HappoAgentLogger()
	key := params["key"]

	var val []byte
	err := db.DB.View(func(tx *bolt.Tx) error {
		bucket := db.MachineStateBucket(tx)
		val = bucket.Get([]byte(key))
		return nil
	})
	if err != nil {
		log.Error(err)
		r.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	r.JSON(http.StatusOK, map[string]string{"machineState": string(val)})
}
