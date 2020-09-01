package pvc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"regexp"
	"strings"
)

type IHandler interface {
	Handle(volId string, tags map[string]string)
}

type HandlerCaller struct {
	client      kubernetes.Interface
	handler     IHandler
	defaultTags map[string]string
}

const (
	SeparatorAnnotation = "volume.beta.kubernetes.io/additional-resource-tags-separator"
	TaglistAnnotation   = "volume.beta.kubernetes.io/additional-resource-tags"
	DefaultSeparator    = ","
)

func (h HandlerCaller) handle(obj interface{}) {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		log.Warnf("Failed to parse pvc handler")
		return
	}
	log.Debugf("Received update event for pvc %s in namespace %s", pvc.Name, pvc.Namespace)

	if !isEBSVolume(pvc) {
		log.Warnf("PVC %s in namespace %s is not an ebs volume, stop handle", pvc.Name, pvc.Namespace)
		return
	}

	volId, vErr := h.receiveVolumeId(pvc.Spec.VolumeName)
	if vErr != nil {
		log.Errorf("Failed to receive EBS volume associated with pvc %s in namespace %s with error %v", pvc.Name, pvc.Namespace, vErr)
		return
	}
	log.Debugf("Received vol id %s", *volId)

	tagSep := loadTagSeparator(pvc)
	tagList := loadTags(pvc, tagSep)
	mergedTagList := mergeTags(h.defaultTags, tagList)
	h.handler.Handle(*volId, mergedTagList)
}

func (h HandlerCaller) receiveVolumeId(volumeName string) (*string, error) {
	awsVolume, errp := h.client.CoreV1().PersistentVolumes().Get(context.TODO(), volumeName, metav1.GetOptions{})
	if errp != nil {
		return nil, errp
	}

	if awsVolume.Spec.PersistentVolumeSource.AWSElasticBlockStore != nil {
		volTag := awsVolume.Spec.PersistentVolumeSource.AWSElasticBlockStore.VolumeID
		r, _ := regexp.Compile(".*?:[\\/]{2,3}.*?\\/(.*)$")
		matches := r.FindStringSubmatch(volTag)
		if matches != nil && len(matches) == 2 {
			return &matches[1], nil
		}
	}

	return nil, fmt.Errorf("couldn't find VolumeId for persistent volume, %s", volumeName)
}

func loadTagSeparator(pvc *v1.PersistentVolumeClaim) string {
	separator := loadAnnotation(pvc, SeparatorAnnotation)
	if separator != nil {
		log.Debugf("Custom separator defined %s", *separator)
		return *separator
	}

	return DefaultSeparator
}

func loadTags(pvc *v1.PersistentVolumeClaim, separator string) map[string]string {
	tagString := loadAnnotation(pvc, TaglistAnnotation)
	if tagString == nil {
		log.Debugf("No additional tags defined")
		return nil
	}

	tagList := strings.Split(*tagString, separator)
	tagMap := map[string]string{}
	for i := range tagList {
		tagEntry := strings.Split(tagList[i], "=")
		if len(tagEntry) != 2 {
			log.Warnf("Ignoring tag '%s' for pvc %s in namespace %s, invalid format!", tagList[i], pvc.Name, pvc.Namespace)
			continue
		}
		key := tagEntry[0]
		val := tagEntry[1]
		tagMap[key] = val
	}

	return tagMap
}

func loadAnnotation(pvc *v1.PersistentVolumeClaim, tag string) *string {
	if pvc.Annotations != nil {
		if val, ok := pvc.Annotations[tag]; ok {
			return &val
		}
	}

	return nil
}

func mergeTags(defaultTags map[string]string, overwriteTags map[string]string) map[string]string {
	mergedTags := map[string]string{}

	if defaultTags != nil {
		for k := range defaultTags {
			mergedTags[k] = defaultTags[k]
		}
	}

	if overwriteTags != nil {
		// overwrite default with overwrite tags
		for k := range overwriteTags {
			mergedTags[k] = overwriteTags[k]
		}
	}

	log.Debugf("Merged tag list: %v", mergedTags)
	return mergedTags
}
