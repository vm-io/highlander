package highlander

import (
	"fmt"
	"sort"
	"github.com/pkg/errors"
)

// New performs initial configuration for the provider
// initializes region, autodiscovers asg name etc.
func New(options AutoScalingGroupOptions) (a *AutoScalingGroup, err error) {
	ec2client, err := NewEc2FromSession(options.AWSSession)
	if err != nil {
		return nil, errors.Wrap(err, "can't create ec2 client")
	}

	metaclient, err := NewMetadataFromSession(options.AWSSession)
	if err != nil {
		return nil, errors.Wrap(err, "can't create metadata client")
	}

	a.Client = &AWS{
		EC2:      ec2client,
		Metadata: metaclient,
	}

	if options.ConfigureFromSelf {
		instance, err := getCurrentInstance(a.Client)
		if err != nil {
			return nil, errors.Wrap(err, "can't retrieve current instance")
		}
		tag, err := tagValue(asgTagName, instance.Tags)
		if err != nil {
			return nil, errors.Wrap(err, "can't retrieve autoscalinggroup tag")
		}
		a.Name = tag
	} else if options.ConfigureFromName != "" {
		a.Name = options.ConfigureFromName
	}

	return
}

// GetMembers returns a list of instances in the autoscaling group
func (a *AutoScalingGroup) GetMembers() ([]AutoScalingGroupMember, error) {
	retval := []AutoScalingGroupMember{}
	instances, err := getRunningInstancesByTag(a.Client, asgTagName, a.Name)
	if err != nil {
		return retval, errors.Wrap(err, "can't retrieve instance tags")
	}

	for _, instance := range instances {
		m := AutoScalingGroupMember{
			Instance: instance,
		}
		retval = append(retval, m)
	}

	sort.Sort(launchTimeSorter(retval))
	return retval, nil
}

// GetLeader finds and returns leader instance in the autoscaling group
// Election rule is to return the oldest running instance
func (a *AutoScalingGroup) GetLeader() (AutoScalingGroupMember, error) {
	members, err := a.GetMembers()
	if err != nil {
		return AutoScalingGroupMember{}, errors.Wrap(err, "error retrieving members")
	}
	if len(members) == 0 {
		return AutoScalingGroupMember{}, fmt.Errorf("no members found, can't elect leader")
	}
	return members[0], nil
}

// GetID returns id for a member. In this case, instance-id
func (m AutoScalingGroupMember) GetID() (string, error) {
	if m.Instance == nil {
		return "", fmt.Errorf("Can't get ID for non-existing instance")
	}
	return *m.Instance.InstanceId, nil
}

// GetName returns a name for a member. In this case, Name tag for instance
func (m AutoScalingGroupMember) GetName() (string, error) {
	if m.Instance == nil {
		return "", fmt.Errorf("Can't get name for non-existing instance")
	}
	return tagValue("Name", m.Instance.Tags)
}

// GetAttribute returns attribute for this instance
func (m AutoScalingGroupMember) GetAttribute(attr string) (interface{}, error) {
	if attr == "instance" {
		return m.Instance, nil
	}
	return nil, fmt.Errorf("Attribute %s not supported by provider", attr)
}
