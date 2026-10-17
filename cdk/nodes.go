package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

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

	privateSubnets := *vpc.PrivateSubnets()
	nodeSG := awsec2.NewSecurityGroup(scope, jsii.String("NodeSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("CKA nodes"),
		AllowAllOutbound: jsii.Bool(true),
	})

	// nodes may talk to each other on any port (simple for a lab)
	nodeSG.AddIngressRule(nodeSG, awsec2.Port_AllTraffic(), jsii.String("node to node"), nil)
	names := []string{"node-1", "node-2", "node-3", "node-4"}

	for i, name := range names {
		awsec2.NewInstance(scope, jsii.String(name), &awsec2.InstanceProps{
			Vpc:           vpc,
			SecurityGroup: nodeSG,
			VpcSubnets:    &awsec2.SubnetSelection{Subnets: &[]awsec2.ISubnet{privateSubnets[i%2]}},
			InstanceType:  awsec2.NewInstanceType(jsii.String(instanceType)),
			MachineImage:  awsec2.MachineImage_LatestAmazonLinux2023(nil),
			Role:          role,
		})
	}
	masterNodeNames := []string{"controlplane-1"}

	for i, name := range masterNodeNames {
		awsec2.NewInstance(scope, jsii.String(name), &awsec2.InstanceProps{
			Vpc:           vpc,
			SecurityGroup: nodeSG,
			VpcSubnets:    &awsec2.SubnetSelection{Subnets: &[]awsec2.ISubnet{privateSubnets[i%2]}},
			InstanceType:  awsec2.NewInstanceType(jsii.String(instanceType)),
			MachineImage:  awsec2.MachineImage_LatestAmazonLinux2023(nil),
			Role:          role,
		})
	}
}
