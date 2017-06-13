package db

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	DB                        *leveldb.DB
	MetricsMaxLifetimeSeconds int64
)

func Open(dbfile string) {
	var err error
	DB, err = leveldb.OpenFile(dbfile, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func Close() {
	var err error
	err = DB.Close()
	if err != nil {
		log.Println(err)
	}
}
