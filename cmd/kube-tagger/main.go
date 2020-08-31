package main

import (
	"flag"
	"github.com/sergiorua/kube-network-flow/pkg/aws"
	"github.com/sergiorua/kube-network-flow/pkg/pvc"
	"github.com/sergiorua/kube-network-flow/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var verbose bool
var kubeconfig string
var defaultTags util.MapFlag
var Version = "development"

func init() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&kubeconfig, "kube-config", "", "Path to a kubeconfig file")
	flag.Var(&defaultTags, "default-tag", "List of default tags with equal as Separator between key and value key=value")
}

func main() {
	flag.Parse()
	log.Infof("Version: %s", Version)

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	kc, err := util.NewKubernetesClient(kubeconfig)
	if err != nil {
		log.Panicf("Failed to create connection to kubernetes api with error %v", err)
	}

	awsC, awsErr := util.NewAWSClient()
	if awsErr != nil {
		log.Panicf("Failed to create AWS Client with error %v", awsErr)
	}

	marker := aws.NewEBSMarker(awsC)
	c, cErr := pvc.NewPVCController(kc, marker, defaultTags.Items)
	if cErr != nil {
		log.Panicf("Failed to create pvc controller with error %v", cErr)
	}

	c.Controller.Run(wait.NeverStop)
}
