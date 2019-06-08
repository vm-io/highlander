package highlander

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AWS is the main container for all AWS related services
type AWS struct {
	EC2      EC2
	Metadata Metadata
}

// EC2 is an abstraction interface around aws-sdk-go ec2
type EC2 interface {
	DescribeInstances(request *ec2.DescribeInstancesInput) ([]*ec2.Instance, error)
}

// Metadata is an abstraction interface around aws-sdk-go ec2metadata
type Metadata interface {
	GetInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error)
}

// Ec2Wrap is a simple wrapper around ec2.EC2
type Ec2Wrap struct {
	*ec2.EC2
}

// MetaWrap is a simple wrapper around ec2metadata.EC2Metadata
type MetaWrap struct {
	*ec2metadata.EC2Metadata
}

// AutoScalingGroup is a highlander provider, implementing
// cluster member discovery using provider.Kind interface
type AutoScalingGroup struct {
	members []*AutoScalingGroupMember
	options AutoScalingGroupOptions
	Name    string
	Client  *AWS
}

// AutoScalingGroupOptions is a list of configuration options
// for AutoScalingGroup highlander provider
type AutoScalingGroupOptions struct {
	ConfigureFromSelf bool
	ConfigureFromName string
	AWSSession        *session.Session
}

// AutoScalingGroupMember is a singular member in the cluster
// implements provider.Member interface
type AutoScalingGroupMember struct {
	Instance *ec2.Instance
}

type launchTimeSorter []AutoScalingGroupMember
