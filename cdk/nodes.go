package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// nodeUserData returns user data (run as root on first boot) that sets the hostname
// and installs containerd configured for the systemd cgroup driver.
func nodeUserData(hostname string) awsec2.UserData {
	userData := awsec2.UserData_ForLinux(nil)
	userData.AddCommands(
		jsii.String(fmt.Sprintf("hostnamectl set-hostname %s", hostname)),
		jsii.String("apt-get update"),
	)
	return userData
}

// ubuntuImage resolves the latest Ubuntu 24.04 LTS AMI from Canonical's public SSM parameter.
func ubuntuImage() awsec2.IMachineImage {
	return awsec2.MachineImage_FromSsmParameter(
		jsii.String("/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id"),
		&awsec2.SsmParameterImageOptions{Os: awsec2.OperatingSystemType_LINUX},
	)
}

// rootVolume returns a 16 GiB encrypted gp3 root disk (/dev/sda1 is the root device of the Ubuntu AMI).
func rootVolume() *[]*awsec2.BlockDevice {
	return &[]*awsec2.BlockDevice{{
		DeviceName: jsii.String("/dev/sda1"),
		Volume: awsec2.BlockDeviceVolume_Ebs(jsii.Number(16), &awsec2.EbsDeviceOptions{
			VolumeType: awsec2.EbsDeviceVolumeType_GP3,
			Encrypted:  jsii.Bool(true),
		}),
	}}
}

// NewNodes creates 4 EC2 instances, 2 in each private subnet.
func NewNodes(scope constructs.Construct, vpc awsec2.Vpc, bucket awss3.Bucket) {
	instanceRoleName := "cka-prep-ec2-role"
	var instanceType string = "t3.medium"

	// SSM role lets us open a shell without SSH or public IPs
	role := awsiam.NewRole(scope, jsii.String(instanceRoleName), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
		},
	})

	bucket.GrantReadWrite(role, nil)

	publicSubnets := *vpc.PublicSubnets()
	nodeSG := awsec2.NewSecurityGroup(scope, jsii.String("NodeSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("CKA nodes"),
		AllowAllOutbound: jsii.Bool(true),
	})

	// nodes may talk to each other on any port (simple for a lab)
	nodeSG.AddIngressRule(nodeSG, awsec2.Port_AllTraffic(), jsii.String("node to node"), nil)

	// construct IDs must be unique per scope, so look the key pair up once and share it
	keyPair := awsec2.KeyPair_FromKeyPairName(scope, jsii.String("CkaPrepKeyPair"), jsii.String("cka-prep"))

	names := []string{"node-1", "node-2", "node-3", "node-4"}

	for i, name := range names {
		awsec2.NewInstance(scope, jsii.String(name), &awsec2.InstanceProps{
			Vpc:                      vpc,
			SecurityGroup:            nodeSG,
			AssociatePublicIpAddress: jsii.Bool(true),
			VpcSubnets:               &awsec2.SubnetSelection{Subnets: &[]awsec2.ISubnet{publicSubnets[i%2]}},
			InstanceType:             awsec2.NewInstanceType(jsii.String(instanceType)),
			MachineImage:             ubuntuImage(),
			Role:                     role,
			BlockDevices:             rootVolume(),
			UserData:                 nodeUserData(name),
			KeyPair:                  keyPair,

			UserDataCausesReplacement: jsii.Bool(true),
		})
	}
	masterNodeNames := []string{"controlplane-1"}

	for i, name := range masterNodeNames {
		awsec2.NewInstance(scope, jsii.String(name), &awsec2.InstanceProps{
			Vpc:                      vpc,
			SecurityGroup:            nodeSG,
			AssociatePublicIpAddress: jsii.Bool(true),
			VpcSubnets:               &awsec2.SubnetSelection{Subnets: &[]awsec2.ISubnet{publicSubnets[i%2]}},
			InstanceType:             awsec2.NewInstanceType(jsii.String(instanceType)),
			MachineImage:             ubuntuImage(),
			Role:                     role,
			BlockDevices:             rootVolume(),
			UserData:                 nodeUserData(name),

			UserDataCausesReplacement: jsii.Bool(true),
		})
	}
}
