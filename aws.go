package highlander

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// This file holds a thin wrapper around aws-sdk-go. It's main purpose
// is to provide an interface that we could mock in testing

const asgTagName = "aws:autoscaling:groupName"

// NewEc2FromSession initializes new ec2wrap given aws session.Session
func NewEc2FromSession(sess *session.Session) (*Ec2Wrap, error) {
	return &Ec2Wrap{
		EC2: ec2.New(sess),
	}, nil
}

// NewMetadataFromSession initializes new metadata client given aws session.Session
func NewMetadataFromSession(sess *session.Session) (*MetaWrap, error) {
	return &MetaWrap{EC2Metadata: ec2metadata.New(sess)}, nil
}

// DescribeInstances returns  list of instances from aws-sdk, handles pagination
func (e *Ec2Wrap) DescribeInstances(request *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	retval := []*ec2.Instance{}

	var nextToken *string

	for {
		response, err := e.EC2.DescribeInstances(request)
		if err != nil {
			return nil, fmt.Errorf("error listing AWS instances: %q", err)
		}

		for _, reservation := range response.Reservations {
			retval = append(retval, reservation.Instances...)
		}

		nextToken = response.NextToken
		if awssdk.StringValue(nextToken) == "" {
			break
		}
		request.NextToken = nextToken
	}
	return retval, nil
}

func getCurrentInstance(client *AWS) (*ec2.Instance, error) {
	doc, err := client.Metadata.GetInstanceIdentityDocument()
	if err != nil {
		return nil, err
	}

	request := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			awssdk.String(doc.InstanceID),
		},
	}
	instances, err := client.EC2.DescribeInstances(request)
	if err != nil {
		return nil, err
	}

	if len(instances) != 1 {
		return nil, fmt.Errorf("Cannot configure AutoScalingGroup provider from self")
	}

	return instances[0], nil
}

func getRunningInstancesByTag(client *AWS, tagName, tagValue string) ([]*ec2.Instance, error) {

	filterName := fmt.Sprintf("tag:%s", tagName)
	request := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			newFilter("instance-state-name", "running"),
			newFilter(filterName, tagValue),
		},
	}
	return client.EC2.DescribeInstances(request)
}

func newFilter(name string, values ...string) *ec2.Filter {
	retval := &ec2.Filter{Name: awssdk.String(name)}
	for _, v := range values {
		retval.Values = append(retval.Values, awssdk.String(v))
	}
	return retval
}

func tagValue(name string, tags []*ec2.Tag) (string, error) {
	for _, tag := range tags {
		if *tag.Key == name {
			return *tag.Value, nil
		}
	}
	return "", fmt.Errorf("Tag %s not found", name)
}

func (s launchTimeSorter) Len() int {
	 return len(s) 
}
func (s launchTimeSorter) Swap(i, j int) { 
	s[i], s[j] = s[j], s[i] 
}

func (s launchTimeSorter) Less(i, j int) bool { 
	return s[i].Instance.LaunchTime.Before(*s[j].Instance.LaunchTime)
 }
