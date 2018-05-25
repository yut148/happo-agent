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
	SvcEC2         ec2iface.EC2API
	SvcAutoscaling autoscalingiface.AutoScalingAPI
}

// NewAWSClient return AWSClient
func NewAWSClient() *AWSClient {
	sess := session.Must(session.NewSession())
	return &AWSClient{
		SvcAutoscaling: autoscaling.New(sess, aws.NewConfig().WithRegion("ap-northeast-1")),
		SvcEC2:         ec2.New(sess, aws.NewConfig().WithRegion("ap-northeast-1")),
	}
}

func (client *AWSClient) describeAutoScalingInstances(autoScalingGroupName string) ([]*ec2.Instance, error) {
	var autoScalingInstances []*ec2.Instance

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(autoScalingGroupName),
		},
	}

	result, err := client.SvcAutoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, err
	}
	if len(result.AutoScalingGroups) < 1 || result.AutoScalingGroups[0].Instances == nil {
		return autoScalingInstances, nil
	}

	var instanceIds []*string
	for _, instance := range result.AutoScalingGroups[0].Instances {
		if *instance.LifecycleState == "InService" {
			instanceIds = append(instanceIds, aws.String(*instance.InstanceId))
		}
	}
	if len(instanceIds) < 1 {
		return autoScalingInstances, nil
	}

	input2 := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	}

	result2, err := client.SvcEC2.DescribeInstances(input2)
	if err != nil {
		return nil, err
	}
	if len(result2.Reservations) < 1 {
		return autoScalingInstances, nil
	}

	for _, r := range result2.Reservations {
		autoScalingInstances = append(autoScalingInstances, r.Instances[0])
	}

	return autoScalingInstances, nil
}
