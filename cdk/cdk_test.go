package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestCdkStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCdkStack(app, "MyStack", nil)

	// THEN
	template := assertions.Template_FromStack(stack, nil)

	template.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::EC2::Subnet"), jsii.Number(4))
	template.ResourceCountIs(jsii.String("AWS::EC2::NatGateway"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("AWS::EC2::Instance"), jsii.Number(5))
	template.ResourceCountIs(jsii.String("AWS::S3::Bucket"), jsii.Number(1))
}
