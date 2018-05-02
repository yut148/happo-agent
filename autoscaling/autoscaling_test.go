package autoscaling

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"testing"

	"github.com/heartbeatsjp/happo-agent/db"
	"github.com/heartbeatsjp/happo-agent/halib"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	leveldbUtil "github.com/syndtr/goleveldb/leveldb/util"
)

const TestConfigFile = "./testdata/autoscaling_test.yaml"
const TestFailConfigFile = "./testdata/autoscaling_test_fail.yaml"
const TestMultiConfigFile = "./testdata/autoscaling_test_multi.yaml"
const TestEmptyConfigFile = "./testdata/autoscaling_test_empty.yaml"
const TestMissingConfigFile = "./testdata/autoscaling_test_missing.yaml"

func TestAutoScaling(t *testing.T) {
	var cases = []struct {
		name     string
		input    string
		expected []struct {
			name  string
			count int
		}
		isNormalTest bool
	}{
		{
			name:  "dummy-prod-ag",
			input: TestConfigFile,
			expected: []struct {
				name  string
				count int
			}{{"dummy-prod-ag", 10}},
			isNormalTest: true,
		},
		{
			name:  "dummy-prod-ag dummy-stg-ag",
			input: TestMultiConfigFile,
			expected: []struct {
				name  string
				count int
			}{{"dummy-prod-ag", 10}, {"dummy-stg-ag", 4}},
			isNormalTest: true,
		},
		{
			name:  "fail-dummy-prod-ag",
			input: TestFailConfigFile,
			expected: []struct {
				name  string
				count int
			}{{"fail-dummy-prod-ag", 10}},
			isNormalTest: true,
		},
		{
			name:  "dummy-empty-ag",
			input: TestEmptyConfigFile,
			expected: []struct {
				name  string
				count int
			}(nil),
			isNormalTest: true,
		},
		{
			name:  "dummy-missing-ag",
			input: TestMissingConfigFile,
			expected: []struct {
				name  string
				count int
			}(nil),
			isNormalTest: false,
		},
	}

	client := &AWSClient{
		svcEC2:         &mockEC2Client{},
		svcAutoscaling: &mockAutoScalingClient{},
	}
	RefreshAutoScalingInstances(client, "dummy-prod-ag", "dummy-prod-app", 10)
	RefreshAutoScalingInstances(client, "fail-dummy-prod-ag", "fail-dummy-prod-app", 10)
	RefreshAutoScalingInstances(client, "dummy-stg-ag", "dummy-stg-app", 4)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			autoscaling, err := AutoScaling(c.input)
			var actual []struct {
				name  string
				count int
			}

			for _, a := range autoscaling {
				actual = append(actual, struct {
					name  string
					count int
				}{name: a.AutoScalingGroupName, count: len(a.InstanceData)})
			}

			if c.isNormalTest {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestSaveAutoScalingConfig(t *testing.T) {
	var cases = []struct {
		name         string
		input1       halib.AutoScalingConfig
		input2       string
		isNormalTest bool
	}{
		{
			name: "single autoscaling group",
			input1: halib.AutoScalingConfig{
				AutoScalings: []struct {
					AutoScalingGroupName string `yaml:"autoscaling_group_name" json:"autoscaling_group_name"`
					AutoScalingCount     int    `yaml:"autoscaling_count" json:"autoscaling_count"`
					HostPrefix           string `yaml:"host_prefix" json:"host_prefix"`
				}{
					{
						AutoScalingGroupName: "dummy-prod-ag",
						AutoScalingCount:     10,
						HostPrefix:           "dummy-prod-app",
					},
				},
			},
			input2:       "./autoscaling_test_save.yaml",
			isNormalTest: true,
		},
	}

	for _, c := range cases {
		defer os.Remove(c.input2)
		t.Run(c.name, func(t *testing.T) {
			err := SaveAutoScalingConfig(c.input1, c.input2)
			if c.isNormalTest {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetAutoScalingConfig(t *testing.T) {
	var cases = []struct {
		name         string
		input        string
		expected     halib.AutoScalingConfig
		isNormalTest bool
	}{
		{
			name:  "single autoscalng group",
			input: TestConfigFile,
			expected: halib.AutoScalingConfig{
				AutoScalings: []struct {
					AutoScalingGroupName string `yaml:"autoscaling_group_name" json:"autoscaling_group_name"`
					AutoScalingCount     int    `yaml:"autoscaling_count" json:"autoscaling_count"`
					HostPrefix           string `yaml:"host_prefix" json:"host_prefix"`
				}{
					{
						AutoScalingGroupName: "dummy-prod-ag",
						AutoScalingCount:     10,
						HostPrefix:           "dummy-prod-app",
					},
				},
			},
			isNormalTest: true,
		},
		{
			name:  "multi autoscaling group",
			input: TestMultiConfigFile,
			expected: halib.AutoScalingConfig{
				AutoScalings: []struct {
					AutoScalingGroupName string `yaml:"autoscaling_group_name" json:"autoscaling_group_name"`
					AutoScalingCount     int    `yaml:"autoscaling_count" json:"autoscaling_count"`
					HostPrefix           string `yaml:"host_prefix" json:"host_prefix"`
				}{
					{
						AutoScalingGroupName: "dummy-prod-ag",
						AutoScalingCount:     10,
						HostPrefix:           "dummy-prod-app",
					},
					{
						AutoScalingGroupName: "dummy-stg-ag",
						AutoScalingCount:     4,
						HostPrefix:           "dummy-stg-app",
					},
				},
			},
			isNormalTest: true,
		},
		{
			name:  "empty config file",
			input: TestEmptyConfigFile,
			expected: halib.AutoScalingConfig{
				AutoScalings: []struct {
					AutoScalingGroupName string `yaml:"autoscaling_group_name" json:"autoscaling_group_name"`
					AutoScalingCount     int    `yaml:"autoscaling_count" json:"autoscaling_count"`
					HostPrefix           string `yaml:"host_prefix" json:"host_prefix"`
				}(nil),
			},
			isNormalTest: true,
		},
		{
			name:  "missing config file",
			input: TestMissingConfigFile,
			expected: halib.AutoScalingConfig{
				AutoScalings: []struct {
					AutoScalingGroupName string `yaml:"autoscaling_group_name" json:"autoscaling_group_name"`
					AutoScalingCount     int    `yaml:"autoscaling_count" json:"autoscaling_count"`
					HostPrefix           string `yaml:"host_prefix" json:"host_prefix"`
				}(nil),
			},
			isNormalTest: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := GetAutoScalingConfig(c.input)
			if c.isNormalTest {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
			assert.Equal(t, c.expected, actual)
		})
	}
}

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI
}

func (m *mockAutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	output := &autoscaling.DescribeAutoScalingGroupsOutput{AutoScalingGroups: []*autoscaling.Group{{}}}
	switch *input.AutoScalingGroupNames[0] {
	case "dummy-prod-ag":
		output.AutoScalingGroups[0].Instances = []*autoscaling.Instance{
			{InstanceId: aws.String("i-aaaaaa"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-bbbbbb"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-cccccc"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-dddddd"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-eeeeee"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-ffffff"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-gggggg"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-hhhhhh"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-iiiiii"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-jjjjjj"), LifecycleState: aws.String("InService")},
		}
	case "fail-dummy-prod-ag":
		output.AutoScalingGroups[0].Instances = []*autoscaling.Instance{
			{InstanceId: aws.String("i-aaaaaa"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-bbbbbb"), LifecycleState: aws.String("Terminating")},
			{InstanceId: aws.String("i-cccccc"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-dddddd"), LifecycleState: aws.String("Terminated")},
			{InstanceId: aws.String("i-eeeeee"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-ffffff"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-gggggg"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-hhhhhh"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-iiiiii"), LifecycleState: aws.String("Pending")},
			{InstanceId: aws.String("i-jjjjjj"), LifecycleState: aws.String("InService")},
		}
	case "dummy-stg-ag":
		output.AutoScalingGroups[0].Instances = []*autoscaling.Instance{
			{InstanceId: aws.String("i-kkkkkk"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-llllll"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-mmmmmm"), LifecycleState: aws.String("InService")},
			{InstanceId: aws.String("i-nnnnnn"), LifecycleState: aws.String("InService")},
		}
	case "allfali-dummy-stg-ag":
		output.AutoScalingGroups[0].Instances = []*autoscaling.Instance{
			{InstanceId: aws.String("i-kkkkkk"), LifecycleState: aws.String("Terminating")},
			{InstanceId: aws.String("i-llllll"), LifecycleState: aws.String("Terminating")},
			{InstanceId: aws.String("i-mmmmmm"), LifecycleState: aws.String("Terminating")},
			{InstanceId: aws.String("i-nnnnnn"), LifecycleState: aws.String("Terminating")},
		}
	case "nil-dummy-stg-ag":
		output.AutoScalingGroups[0].Instances = []*autoscaling.Instance(nil)
	}
	return output, nil
}

type mockEC2Client struct {
	ec2iface.EC2API
}

func (m *mockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{}}
	reservations := []*ec2.Reservation{
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-aaaaaa"), PrivateIpAddress: aws.String("192.0.2.11")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-bbbbbb"), PrivateIpAddress: aws.String("192.0.2.12")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-cccccc"), PrivateIpAddress: aws.String("192.0.2.13")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-dddddd"), PrivateIpAddress: aws.String("192.0.2.14")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-eeeeee"), PrivateIpAddress: aws.String("192.0.2.15")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-ffffff"), PrivateIpAddress: aws.String("192.0.2.16")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-gggggg"), PrivateIpAddress: aws.String("192.0.2.17")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-hhhhhh"), PrivateIpAddress: aws.String("192.0.2.18")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-iiiiii"), PrivateIpAddress: aws.String("192.0.2.19")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-jjjjjj"), PrivateIpAddress: aws.String("192.0.2.20")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-kkkkkk"), PrivateIpAddress: aws.String("192.0.2.21")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-llllll"), PrivateIpAddress: aws.String("192.0.2.22")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-mmmmmm"), PrivateIpAddress: aws.String("192.0.2.23")}}},
		{Instances: []*ec2.Instance{{InstanceId: aws.String("i-nnnnnn"), PrivateIpAddress: aws.String("192.0.2.24")}}},
	}

	for _, instanceID := range input.InstanceIds {
		for _, r := range reservations {
			if *instanceID == *r.Instances[0].InstanceId {
				output.Reservations = append(output.Reservations, r)
			}
		}
	}

	return output, nil
}

func TestRefreshAutoScalingInstances(t *testing.T) {
	var cases = []struct {
		name     string
		input1   string
		input2   string
		input3   int
		expected []halib.InstanceData
	}{
		{
			name:   "dummy-prod-ag",
			input1: "dummy-prod-ag",
			input2: "dummy-prod-app",
			input3: 10,
			expected: []halib.InstanceData{
				{
					InstanceID: "i-aaaaaa",
					IP:         "192.0.2.11",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-jjjjjj",
					IP:         "192.0.2.20",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-bbbbbb",
					IP:         "192.0.2.12",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-cccccc",
					IP:         "192.0.2.13",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-dddddd",
					IP:         "192.0.2.14",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-eeeeee",
					IP:         "192.0.2.15",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-ffffff",
					IP:         "192.0.2.16",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-gggggg",
					IP:         "192.0.2.17",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-hhhhhh",
					IP:         "192.0.2.18",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-iiiiii",
					IP:         "192.0.2.19",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
			},
		},
		{
			name:   "fail-dummy-prod-ag",
			input1: "fail-dummy-prod-ag",
			input2: "fail-dummy-prod-app",
			input3: 10,
			expected: []halib.InstanceData{
				{
					InstanceID: "i-aaaaaa",
					IP:         "192.0.2.11",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
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
				},
				{
					InstanceID: "i-cccccc",
					IP:         "192.0.2.13",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-eeeeee",
					IP:         "192.0.2.15",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-ffffff",
					IP:         "192.0.2.16",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-gggggg",
					IP:         "192.0.2.17",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-hhhhhh",
					IP:         "192.0.2.18",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-jjjjjj",
					IP:         "192.0.2.20",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
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
				},
				{
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
				},
			},
		},
		{
			name:   "dummy-stg-ag",
			input1: "dummy-stg-ag",
			input2: "dummy-stg-app",
			input3: 4,
			expected: []halib.InstanceData{
				{
					InstanceID: "i-kkkkkk",
					IP:         "192.0.2.21",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-llllll",
					IP:         "192.0.2.22",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-mmmmmm",
					IP:         "192.0.2.23",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
				{
					InstanceID: "i-nnnnnn",
					IP:         "192.0.2.24",
					MetricPlugins: []struct {
						PluginName   string `json:"plugin_name"`
						PluginOption string `json:"plugin_option"`
					}{
						{
							PluginName:   "",
							PluginOption: "",
						},
					},
				},
			},
		},
		{
			name:   "allfail-dummy-stg-ag",
			input1: "allfail-dummy-stg-ag",
			input2: "allfail-dummy-stg-app",
			input3: 4,
			expected: []halib.InstanceData{
				{
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
				},
				{
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
				},
				{
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
				},
				{
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
				},
			},
		},
		{
			name:   "nill-dummy-stg-ag",
			input1: "nill-dummy-stg-ag",
			input2: "nill-dummy-stg-app",
			input3: 4,
			expected: []halib.InstanceData{
				{
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
				},
				{
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
				},
				{
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
				},
				{
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
				},
			},
		},
		{
			name:   "missing instance",
			input1: "dummy-missing-ag",
			input2: "dummy-missing-app",
			input3: 4,
			expected: []halib.InstanceData{
				{
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
				},
				{
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
				},
				{
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
				},
				{
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
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := &AWSClient{
				svcEC2:         &mockEC2Client{},
				svcAutoscaling: &mockAutoScalingClient{},
			}
			err := RefreshAutoScalingInstances(client, c.input1, c.input2, c.input3)
			assert.Nil(t, err)

			iter := db.DB.NewIterator(
				leveldbUtil.BytesPrefix(
					[]byte(fmt.Sprintf("ag-%s-", c.input2)),
				),
				nil,
			)
			var actual []halib.InstanceData
			for iter.Next() {
				value := iter.Value()

				var instanceData halib.InstanceData
				dec := gob.NewDecoder(bytes.NewReader(value))
				dec.Decode(&instanceData)
				actual = append(actual, instanceData)
			}
			iter.Release()

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestMain(m *testing.M) {
	//Mock
	DB, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		os.Exit(1)
	}
	db.DB = DB
	os.Exit(m.Run())

	db.DB.Close()
}
