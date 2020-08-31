package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewAWSClient() (*ec2.EC2, error) {
	awsConfig := &aws.Config{}
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.New(awsSession, awsConfig)
	return ec2Client, nil
}
