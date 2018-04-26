package autoscaling

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// AWSClient allows you to get the list of IP addresses of instanes of an Auto Scaling group
type AWSClient struct {
	svcEC2         ec2iface.EC2API
	svcAutoscaling autoscalingiface.AutoScalingAPI
}

// NewAWSClient return AWSClient
func NewAWSClient() *AWSClient {
	sess := session.Must(session.NewSession())
	return &AWSClient{
		svcAutoscaling: autoscaling.New(sess, aws.NewConfig().WithRegion("ap-northeast-1")),
		svcEC2:         ec2.New(sess, aws.NewConfig().WithRegion("ap-northeast-1")),
	}
}

func (client *AWSClient) describeAutoScalingInstances(autoScalingGroupName string) (*ec2.DescribeInstancesOutput, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(autoScalingGroupName),
		},
	}

	result, err := client.svcAutoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}

	var instances []*string
	for _, instance := range result.AutoScalingGroups[0].Instances {
		instances = append(instances, aws.String(*instance.InstanceId))
	}

	input2 := &ec2.DescribeInstancesInput{
		InstanceIds: instances,
	}

	return client.svcEC2.DescribeInstances(input2)
}
