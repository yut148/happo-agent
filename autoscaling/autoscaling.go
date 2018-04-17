package autoscaling

import (
	"io/ioutil"
	"os"

	"github.com/heartbeatsjp/happo-agent/halib"

	yaml "gopkg.in/yaml.v2"
)

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
