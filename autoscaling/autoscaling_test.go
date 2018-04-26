package autoscaling

import (
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
)

const TestConfigFile = "./autoscaling_test.yaml"
const TestMultiConfigFile = "./autoscaling_test_multi.yaml"
const TestEmptyConfigFile = "./autoscaling_test_empty.yaml"

// func TestAutoScaling(t *testing.T) {
// 	var cases = []struct {
// 		name         string
// 		input        string
// 		expected     []halib.AutoScalingData
// 		isNormalTest bool
// 	}{
// 		{
// 			name:  "default",
// 			input: TestConfigFile,
// 			expected: []halib.AutoScalingData{
// 				{
// 					AutoScalingGroupName: "dummy-prod-ag",
// 					InstanceData: map[string]halib.InstanceData{
// 						"dummy-prod-app-1": {
// 							IP:         "",
// 							InstanceID: "",
// 							MetricPlugins: []struct {
// 								PluginName   string `json:"plugin_name"`
// 								PluginOption string `json:"plugin_option"`
// 							}{},
// 						},
// 					},
// 				},
// 			},
// 			isNormalTest: true,
// 		},
// 	}
//
// 	for _, c := range cases {
// 		t.Run(c.name, func(t *testing.T) {
// 			_, err := AutoScaling(c.input)
// 			if c.isNormalTest {
// 				assert.Nil(t, err)
// 			} else {
// 				assert.NotNil(t, err)
// 			}
// 		})
// 	}
// }

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
			input: "./autoscaling_dummy.yaml",
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
	}

	for _, instanceID := range input.InstanceIds {
		for _, r := range reservations {
			if instanceID == r.Instances[0].InstanceId {
				output.Reservations = append(output.Reservations, r)
			}
		}
	}

	return output, nil
}

func TestRefreshAutoScalingInstances(t *testing.T) {
	var cases = []struct {
		name         string
		input1       string
		input2       string
		input3       int
		isNormalTest bool
	}{
		{
			name:         "default",
			input1:       "dummy-prod-ag",
			input2:       "dummy-prod-app",
			input3:       10,
			isNormalTest: true,
		},
		{
			name:         "missing instance",
			input1:       "dummy-missing-ag",
			input2:       "dummy-missing-app",
			input3:       10,
			isNormalTest: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			client := &AWSClient{
				svcEC2:         &mockEC2Client{},
				svcAutoscaling: &mockAutoScalingClient{},
			}
			err := RefreshAutoScalingInstances(client, c.input1, c.input2, c.input3)
			if c.isNormalTest {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
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
