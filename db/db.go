package db

import (
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/heartbeatsjp/happo-agent/util"
)

var (
	// DB is bolt.DB
	DB *bolt.DB
	// MetricsMaxLifetimeSeconds is as variable name
	MetricsMaxLifetimeSeconds int64
	// MachineStateMaxLifetimeSeconds is as variable name
	MachineStateMaxLifetimeSeconds int64
	metricBucketName               []byte
	machieStateBucketName          []byte
)

func init() {
	MetricsMaxLifetimeSeconds = 7 * 86400      //default is 7 days
	MachineStateMaxLifetimeSeconds = 3 * 86400 //default is 3 days
	metricBucketName = []byte("Metric")
	machieStateBucketName = []byte("MachineState")
}

// Open open leveldb file
func Open(dbfile string) {
	var err error
	log := util.HappoAgentLogger()

	if fsinfo, err := os.Stat(dbfile); err != nil {
		err = os.MkdirAll(dbfile, 0755)
		if err != nil {
			log.Fatalln("mkdir dbfile failed. ", err)
		}
	} else if !fsinfo.IsDir() {
		log.Fatalln("cannot create dbfile/data.db. dbfile path is already used as directory")
	}

	dbfile = filepath.Join(dbfile, "data.db")
	DB, err = bolt.Open(dbfile, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	err = DB.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists(metricBucketName)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(machieStateBucketName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}

// Close close leveldb file
func Close() {
	log := util.HappoAgentLogger()
	var err error
	err = DB.Close()
	if err != nil {
		log.Error(err)
	}
}

// MetricBucket returns bucket for metrics
func MetricBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(metricBucketName)
}

// MachineStateBucket returns bucket for metrics
func MachineStateBucket(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket(machieStateBucketName)
}

// TimeToKey returns key
func TimeToKey(t time.Time) []byte {
	return []byte(t.Format(time.RFC3339))
}

// KeyToUnixtime returns unixtime
// key format is RFC3339
func KeyToUnixtime(key []byte) int64 {
	log := util.HappoAgentLogger()
	t, err := time.Parse(time.RFC3339, string(key))
	if err != nil {
		log.Errorf("key %v parse failed. %v", string(key), err)
	}
	return t.Unix()
}
