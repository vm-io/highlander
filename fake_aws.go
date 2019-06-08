package highlander

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// FakeEC2 is a mock implementation for EC2
type FakeEC2 struct {
	Instances []*ec2.Instance
}

// FakeMetadata is a mock implementation for Metadata
type FakeMetadata struct {
	document *ec2metadata.EC2InstanceIdentityDocument
}

// DescribeInstances returns  list of instances from aws-sdk, handles pagination
func (e *FakeEC2) DescribeInstances(request *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	retval := e.Instances

	for _, filter := range request.Filters {
		retval = e.applyFilter(filter, retval)
	}
	if request.InstanceIds != nil {
		retval = e.applyInstanceFilter(request.InstanceIds, retval)
	}
	return retval, nil
}

func (e *FakeEC2) applyFilter(filter *ec2.Filter, instances []*ec2.Instance) []*ec2.Instance {
	retval := []*ec2.Instance{}
	filterName := *filter.Name

	for _, instance := range instances {
		if filterName == "instance-state-name" {
			for _, value := range filter.Values {
				if *instance.State.Name == *value {
					retval = append(retval, instance)
					break
				}
			}
		}
		if strings.HasPrefix(filterName, "tag:") {
			tagName := filterName[4:]
			ok, _ := tagValue(tagName, instance.Tags)
			for _, val := range filter.Values {
				if *val == ok {
					retval = append(retval, instance)
					break
				}
			}
		}
	}

	return retval
}

func (e *FakeEC2) applyInstanceFilter(ids []*string, instances []*ec2.Instance) []*ec2.Instance {
	retval := []*ec2.Instance{}

	for _, instanceID := range ids {
		for _, instance := range instances {
			if *instance.InstanceId == *instanceID {
				retval = append(retval, instance)
			}
		}

	}
	return retval
}

// GetInstanceIdentityDocument is a mock for testing
func (m *FakeMetadata) GetInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error) {
	return *m.document, nil
}
