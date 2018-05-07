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
		autoScalingData.InstanceData = map[string]halib.InstanceData{}

		iter := transaction.NewIterator(
			leveldbUtil.BytesPrefix(
				[]byte(fmt.Sprintf("ag-%s-", a.HostPrefix)),
			),
			nil,
		)
		for iter.Next() {
			var instanceData halib.InstanceData
			alias := strings.TrimPrefix(string(iter.Key()), "ag-")
			value := iter.Value()
			dec := gob.NewDecoder(bytes.NewReader(value))
			dec.Decode(&instanceData)
			autoScalingData.InstanceData[alias] = instanceData
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

func makeRegisteredInstances(transaction *leveldb.Transaction, hostPrefix string) map[string]halib.InstanceData {
	registeredInstances := map[string]halib.InstanceData{}

	iter := transaction.NewIterator(
		leveldbUtil.BytesPrefix(
			[]byte(fmt.Sprintf("ag-%s-", hostPrefix)),
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
	registeredInstances := makeRegisteredInstances(transaction, hostPrefix)

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
				instanceData.MetricPlugins = []struct {
					PluginName   string `json:"plugin_name"`
					PluginOption string `json:"plugin_option"`
				}{
					{
						PluginName:   "",
						PluginOption: "",
					},
				}
				newInstances = append(newInstances, instanceData)
			}
		}

		for _, instance := range newInstances {
			for i := 0; i < autoscalingCount; i++ {
				key := fmt.Sprintf("ag-%s-%d", hostPrefix, i+1)
				if _, ok := actualInstances[key]; !ok {
					if _, ok := registeredInstances[key]; ok {
						instance.MetricPlugins = registeredInstances[key].MetricPlugins
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
			InstanceID: "",
			IP:         "",
			MetricPlugins: []struct {
				PluginName   string `json:"plugin_name"`
				PluginOption string `json:"plugin_option"`
			}{
				{
					PluginName:   "",
					PluginOption: "",
				},
			},
		}
		key := fmt.Sprintf("ag-%s-%d", hostPrefix, i+1)
		if _, ok := actualInstances[key]; !ok {
			if _, ok := registeredInstances[key]; ok {
				emptyInstance.MetricPlugins = registeredInstances[key].MetricPlugins
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
