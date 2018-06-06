package autoscaling

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"
	"github.com/heartbeatsjp/happo-agent/util"
	"github.com/syndtr/goleveldb/leveldb"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"

	yaml "gopkg.in/yaml.v2"
)

// AutoScaling list autoscaling instances
func AutoScaling(configPath string) ([]halib.AutoScalingData, error) {
	log := util.HappoAgentLogger()
	var autoScaling []halib.AutoScalingData

	autoScalingList, err := GetAutoScalingConfig(configPath)
	if err != nil {
		return autoScaling, err
	}

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return autoScaling, err
	}

	for _, a := range autoScalingList.AutoScalings {
		var autoScalingData halib.AutoScalingData
		autoScalingData.AutoScalingGroupName = a.AutoScalingGroupName
		autoScalingData.Instances = []struct {
			Alias        string             `json:"alias"`
			InstanceData halib.InstanceData `json:"instance_data"`
		}{}

		iter := transaction.NewIterator(
			leveldbUtil.BytesPrefix(
				[]byte(fmt.Sprintf("ag-%s-%s-", a.AutoScalingGroupName, a.HostPrefix)),
			),
			nil,
		)
		for iter.Next() {
			var instanceData halib.InstanceData
			alias := strings.TrimPrefix(string(iter.Key()), "ag-")
			value := iter.Value()
			dec := gob.NewDecoder(bytes.NewReader(value))
			dec.Decode(&instanceData)
			autoScalingData.Instances = append(autoScalingData.Instances, struct {
				Alias        string             `json:"alias"`
				InstanceData halib.InstanceData `json:"instance_data"`
			}{
				Alias:        alias,
				InstanceData: instanceData,
			})
		}
		autoScaling = append(autoScaling, autoScalingData)
		iter.Release()
	}

	transaction.Discard()

	return autoScaling, nil
}

// SaveAutoScalingConfig save autoscaling config to config file
func SaveAutoScalingConfig(config halib.AutoScalingConfig, configFile string) error {
	buf, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFile, buf, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// GetAutoScalingConfig returns autoscaling config file
func GetAutoScalingConfig(configFile string) (halib.AutoScalingConfig, error) {
	var autoscalingConfig halib.AutoScalingConfig

	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return autoscalingConfig, err
	}
	err = yaml.Unmarshal(buf, &autoscalingConfig)
	if err != nil {
		return autoscalingConfig, err
	}

	return autoscalingConfig, nil
}

func getInstanceData(transaction *leveldb.Transaction, instanceID string) ([]byte, halib.InstanceData) {
	var key []byte
	var instanceData halib.InstanceData

	iter := transaction.NewIterator(leveldbUtil.BytesPrefix([]byte("ag-")), nil)
	for iter.Next() {
		var d halib.InstanceData
		value := iter.Value()

		dec := gob.NewDecoder(bytes.NewReader(value))
		dec.Decode(&d)
		if d.InstanceID == instanceID {
			key = iter.Key()
			instanceData = d
			break
		}
	}

	return key, instanceData
}

func getEmptyAlias(transaction *leveldb.Transaction, autoScalingGroupName, hostPrefix string) (string, halib.InstanceData) {
	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix(
			[]byte(fmt.Sprintf("ag-%s-%s-", autoScalingGroupName, hostPrefix)),
		),
		nil)

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var instanceData halib.InstanceData
		dec := gob.NewDecoder(bytes.NewReader(value))
		dec.Decode(&instanceData)
		if instanceData.InstanceID == "" {
			return strings.TrimPrefix(string(key), "ag-"), instanceData
		}
	}
	iter.Release()
	return "", halib.InstanceData{}
}

// RegisterAutoScalingInstance register autoscaling instance to dbms
func RegisterAutoScalingInstance(autoScalingGroupName, hostPrefix, instanceID, ip string) (string, halib.InstanceData, error) {
	log := util.HappoAgentLogger()

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return "", halib.InstanceData{}, err
	}

	registeredInstances := makeRegisteredInstances(transaction, autoScalingGroupName, hostPrefix)
	for _, registeredInstance := range registeredInstances {
		if instanceID == registeredInstance.InstanceID {
			transaction.Discard()
			return "", halib.InstanceData{}, fmt.Errorf("already registered")
		}
	}

	newAlias, newInstanceData := getEmptyAlias(transaction, autoScalingGroupName, hostPrefix)
	if newAlias == "" {
		transaction.Discard()
		return "", halib.InstanceData{}, fmt.Errorf("can't find empty alias from %s", autoScalingGroupName)
	}

	newInstanceData.InstanceID = instanceID
	newInstanceData.IP = ip

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(newInstanceData)
	transaction.Put(
		[]byte(fmt.Sprintf("ag-%s", newAlias)),
		b.Bytes(),
		nil)

	if err := transaction.Commit(); err != nil {
		log.Error(err)
		return "", halib.InstanceData{}, err
	}

	return newAlias, newInstanceData, nil
}

// DeregisterAutoScalingInstance deregister autoscaling instance from dbms
func DeregisterAutoScalingInstance(instanceID string) error {
	log := util.HappoAgentLogger()

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return err
	}

	key, instanceData := getInstanceData(transaction, instanceID)
	if key == nil {
		transaction.Discard()
		err := fmt.Errorf("%s is not registered", instanceID)
		log.Error(err)
		return err
	}

	err = transaction.Delete(key, nil)
	if err != nil {
		transaction.Discard()
		log.Error(err)
		return err
	}

	instanceData.InstanceID = ""
	instanceData.IP = ""
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(instanceData)
	transaction.Put(
		key,
		b.Bytes(),
		nil,
	)

	err = transaction.Commit()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func makeRegisteredInstances(transaction *leveldb.Transaction, autoScalingGroupName, hostPrefix string) map[string]halib.InstanceData {
	registeredInstances := map[string]halib.InstanceData{}

	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix(
			[]byte(fmt.Sprintf("ag-%s-%s-", autoScalingGroupName, hostPrefix)),
		),
		nil,
	)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		var instanceData halib.InstanceData
		dec := gob.NewDecoder(bytes.NewReader(value))
		dec.Decode(&instanceData)
		registeredInstances[string(key)] = instanceData
	}
	iter.Release()
	return registeredInstances
}

func checkRegistered(instance *ec2.Instance, registeredInstances map[string]halib.InstanceData) (bool, string) {
	isRegistered := false
	var registeredKey string
	for key, registerdInstance := range registeredInstances {
		if *instance.InstanceId == registerdInstance.InstanceID {
			registeredKey = key
			isRegistered = true
			break
		}
	}
	return isRegistered, registeredKey
}

// RefreshAutoScalingInstances refresh alias maps
func RefreshAutoScalingInstances(client *AWSClient, autoScalingGroupName, hostPrefix string, autoscalingCount int) error {
	log := util.HappoAgentLogger()

	autoScalingInstances, err := client.describeAutoScalingInstances(autoScalingGroupName)
	if err != nil {
		log.Error(err)
		return err
	}

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return err
	}

	// registerdInstance has already been registered instances with dbms
	registeredInstances := makeRegisteredInstances(transaction, autoScalingGroupName, hostPrefix)

	// init dbms
	for key := range registeredInstances {
		transaction.Delete([]byte(key), nil)
	}

	// newInstances has not been registered with dbms
	newInstances := []halib.InstanceData{}

	// actualInstances will be registered to the dbms
	actualInstances := map[string]halib.InstanceData{}

	// if there are autoscaling instances,
	// put in actualInstances at same key an instances of registered with dbms in autoscaling instances
	// after there, put in actualInstances at empty key an instances of not registered with dbms in autoscaling instances
	if len(autoScalingInstances) > 0 {
		for _, autoScalingInstance := range autoScalingInstances {
			if isRegistered, key := checkRegistered(autoScalingInstance, registeredInstances); isRegistered {
				actualInstances[key] = registeredInstances[key]
			} else {
				var instanceData halib.InstanceData
				instanceData.InstanceID = *autoScalingInstance.InstanceId
				instanceData.IP = *autoScalingInstance.PrivateIpAddress
				instanceData.MetricConfig = halib.MetricConfig{}
				newInstances = append(newInstances, instanceData)
			}
		}

		for _, instance := range newInstances {
			for i := 0; i < autoscalingCount; i++ {
				key := fmt.Sprintf("ag-%s-%s-%d", autoScalingGroupName, hostPrefix, i+1)
				if _, ok := actualInstances[key]; !ok {
					if _, ok := registeredInstances[key]; ok {
						instance.MetricConfig = registeredInstances[key].MetricConfig
					}
					actualInstances[key] = instance
					break
				}
			}
		}
	}

	// fill actualInstances with emptyInstance
	for i := 0; i < autoscalingCount; i++ {
		emptyInstance := halib.InstanceData{
			InstanceID:   "",
			IP:           "",
			MetricConfig: halib.MetricConfig{},
		}
		key := fmt.Sprintf("ag-%s-%s-%d", autoScalingGroupName, hostPrefix, i+1)
		if _, ok := actualInstances[key]; !ok {
			if _, ok := registeredInstances[key]; ok {
				emptyInstance.MetricConfig = registeredInstances[key].MetricConfig
			}
			actualInstances[key] = emptyInstance
		}
	}

	// actualInstances register to dbms
	batch := new(leveldb.Batch)
	for key, value := range actualInstances {
		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		err = enc.Encode(value)
		batch.Put(
			[]byte(key),
			b.Bytes(),
		)
	}
	err = transaction.Write(batch, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	err = transaction.Commit()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// DeleteAutoScaling delete autoscaling instances data
func DeleteAutoScaling(autoScalingGroupName string) error {
	log := util.HappoAgentLogger()

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return err
	}

	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix(
			[]byte(fmt.Sprintf("ag-%s-", autoScalingGroupName)),
		),
		nil,
	)
	batch := new(leveldb.Batch)
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
	}
	iter.Release()

	err = transaction.Write(batch, nil)
	if err != nil {
		transaction.Discard()
		log.Error(err)
		return err
	}

	err = transaction.Commit()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// SaveAliasMetricConfig save metric config of autoscaling instance to dbms
func SaveAliasMetricConfig(alias string, metricConfig halib.MetricConfig) error {
	log := util.HappoAgentLogger()

	transaction, err := db.DB.OpenTransaction()
	if err != nil {
		log.Error(err)
		return err
	}

	key := []byte(fmt.Sprintf("ag-%s", alias))

	value, err := transaction.Get(key, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	var instanceData halib.InstanceData
	dec := gob.NewDecoder(bytes.NewReader(value))
	if err := dec.Decode(&instanceData); err != nil {
		transaction.Discard()
		log.Error(err)
		return err
	}

	instanceData.MetricConfig = metricConfig

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(instanceData); err != nil {
		transaction.Discard()
		log.Error(err)
		return err
	}

	if err := transaction.Put(key, b.Bytes(), nil); err != nil {
		transaction.Discard()
		log.Error(err)
		return err
	}

	if err := transaction.Commit(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// AliasToIP resolve autoscaling instance private ip address
func AliasToIP(alias string) (string, error) {
	value, err := db.DB.Get([]byte(fmt.Sprintf("ag-%s", alias)), nil)
	if err != nil {
		return "", err
	}
	var instanceData halib.InstanceData
	dec := gob.NewDecoder(bytes.NewReader(value))
	if err := dec.Decode(&instanceData); err != nil {
		return "", err
	}
	return instanceData.IP, nil
}
