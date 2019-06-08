package highlander

import (
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestGetMembers(t *testing.T) {
	cli := &AWS{
		EC2:      newFakeEC2(),
		Metadata: newFakeMetadata(),
	}
	group := &AutoScalingGroup{
		Client: cli,
		Name:   "asg",
	}
	members, err := group.GetMembers()
	if err != nil {
		t.Error(err)
	}
	if len(members) != 2 {
		t.Errorf("Unexpected member count %d", len(members))
	}

}

func TestNoInstances(t *testing.T) {
	cli := &AWS{
		EC2:      &FakeEC2{},
		Metadata: newFakeMetadata(),
	}
	group := &AutoScalingGroup{
		Client: cli,
		Name:   "asg",
	}

	members, err := group.GetMembers()
	if err != nil {
		t.Error(err)
	}
	if len(members) != 0 {
		t.Errorf("Unexpected member count %d", len(members))
	}

	_, err = group.GetLeader()
	if err == nil {
		t.Errorf("No leader found, should raise error")
	}
}

func TestGetLeader(t *testing.T) {
	cli := &AWS{
		EC2:      newFakeEC2(),
		Metadata: newFakeMetadata(),
	}
	group := &AutoScalingGroup{
		Client: cli,
		Name:   "asg",
	}

	leader, err := group.GetLeader()
	if err != nil {
		t.Error(err)
	}
	id, err := leader.GetID()
	if err != nil {
		t.Error(err)
	}
	if id != "i-something3" {
		t.Errorf("Unexpected leader: %s", id)
	}

	name, err := leader.GetName()
	if err != nil {
		t.Error(err)
	}
	if name != "asg3" {
		t.Errorf("Unexpected leader name: %s", name)
	}
}

func TestNullMember(t *testing.T) {
	member := AutoScalingGroupMember{}

	_, err := member.GetID()
	if err == nil {
		t.Errorf("Unexpected id for empty member. Should raise error, but got nil")
	}

	_, err = member.GetName()
	if err == nil {
		t.Errorf("Unexpected name for empty member. Should raise error, but got nil")
	}
}

type instanceInfo struct {
	ID               string
	AvailabilityZone string
	PrivateDNS       string
	PrivateIP        string
	PublicIP         string
	State            string
	Tags             map[string]string
	LaunchTime       string
}

func newFakeEC2() *FakeEC2 {
	return &FakeEC2{
		Instances: []*ec2.Instance{
			newFakeInstance(instanceInfo{
				ID:               "i-something1",
				AvailabilityZone: "us-east-1a",
				PrivateDNS:       "ip-172-20-0-1.ec2.internal",
				PrivateIP:        "172.20.0.1",
				Tags: map[string]string{
					"Name":                      "asg1",
					"aws:autoscaling:groupName": "asg1",
				},
				State:      "running",
				LaunchTime: "2007-01-02T15:04:05Z",
			}),
			newFakeInstance(instanceInfo{
				ID:               "i-something2",
				AvailabilityZone: "us-east-1a",
				PrivateDNS:       "ip-172-20-0-2.ec2.internal",
				PrivateIP:        "172.20.0.2",
				Tags: map[string]string{
					"Name":                      "asg2",
					"aws:autoscaling:groupName": "asg",
				},
				State:      "running",
				LaunchTime: "2006-02-02T15:04:05Z",
			}),
			newFakeInstance(instanceInfo{
				ID:               "i-something3",
				AvailabilityZone: "us-east-1a",
				PrivateDNS:       "ip-172-20-0-3.ec2.internal",
				PrivateIP:        "172.20.0.3",
				Tags: map[string]string{
					"Name":                      "asg3",
					"aws:autoscaling:groupName": "asg",
				},
				State:      "running",
				LaunchTime: "2006-01-03T15:04:05Z",
			}),
			newFakeInstance(instanceInfo{
				ID:               "i-something4",
				AvailabilityZone: "us-east-1a",
				PrivateDNS:       "ip-172-20-0-4.ec2.internal",
				PrivateIP:        "172.20.0.4",
				Tags: map[string]string{
					"Name": "asg4",
				},
				State:      "running",
				LaunchTime: "2006-01-02T15:04:05Z",
			}),
			newFakeInstance(instanceInfo{
				ID:               "i-something5",
				AvailabilityZone: "us-east-1a",
				PrivateDNS:       "ip-172-20-0-5.ec2.internal",
				PrivateIP:        "172.20.0.5",
				Tags: map[string]string{
					"Name":                      "asg5",
					"aws:autoscaling:groupName": "asg",
				},
				State:      "pending",
				LaunchTime: "2012-01-02T15:04:05Z",
			}),
		},
	}
}

func newFakeMetadata() *FakeMetadata {
	return &FakeMetadata{
		document: &ec2metadata.EC2InstanceIdentityDocument{
			AvailabilityZone: "us-east-1a",
			PrivateIP:        "172.20.0.1",
			Region:           "us-east-1",
		},
	}
}

func newFakeInstance(info instanceInfo) *ec2.Instance {
	retval := &ec2.Instance{
		InstanceId: awssdk.String(info.ID),
		Placement: &ec2.Placement{
			AvailabilityZone: awssdk.String(info.AvailabilityZone),
		},
		PrivateDnsName:   awssdk.String(info.PrivateDNS),
		PublicIpAddress:  awssdk.String(info.PublicIP),
		PrivateIpAddress: awssdk.String(info.PrivateIP),
		State: &ec2.InstanceState{
			Name: awssdk.String(info.State),
		},
	}

	for tag, value := range info.Tags {
		t := &ec2.Tag{
			Key:   awssdk.String(tag),
			Value: awssdk.String(value),
		}
		retval.Tags = append(retval.Tags, t)
	}

	lt, _ := time.Parse(time.RFC3339, info.LaunchTime)
	retval.LaunchTime = &lt
	return retval
}
