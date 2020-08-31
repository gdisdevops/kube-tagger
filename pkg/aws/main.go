package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	log "github.com/sirupsen/logrus"
	"strings"
)

type EBSMarker struct {
	ec2Client ec2iface.EC2API
}

func NewEBSMarker(client *ec2.EC2) EBSMarker {
	return EBSMarker{
		ec2Client: client,
	}
}

func (e EBSMarker) Handle(volId string, tags map[string]string) {
	var tagList []*ec2.Tag

	for k := range tags {
		if strings.ToLower(k) == "name" || strings.ToLower(k) == "KubernetesCluster" || strings.HasPrefix(strings.ToLower(k), "kubernetes.io/") {
			log.Warnf("Tag name is in a list of blacklisted names %s", k)
			continue
		}

		tag := ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(tags[k]),
		}
		tagList = append(tagList, &tag)
	}

	createInput := ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(volId),
		},
		Tags: tagList,
	}

	_, err := e.ec2Client.CreateTags(&createInput)
	if err != nil {
		log.Errorf("Failed to add tags to volume %s with error %v", volId, err)
	}
}
