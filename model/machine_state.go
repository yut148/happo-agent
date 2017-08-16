package model

import (
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/util"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"
)

// ListMachieState returns saved machine states
func ListMachieState(r render.Render) {
	log := util.HappoAgentLogger()
	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Println(err)
		r.JSON(http.StatusNoContent, map[string]string{"error": err.Error()})
		return
	}

	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix([]byte("s-")),
		nil)
	var keys []string
	for iter.Next() {
		keys = append(keys, string(iter.Key()))
	}
	iter.Release()
	transaction.Discard()

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

	val, err := db.DB.Get([]byte(key), nil)

	if err != nil {
		log.Println(err)
		r.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	r.JSON(http.StatusOK, map[string]string{"machineState": string(val)})
}
