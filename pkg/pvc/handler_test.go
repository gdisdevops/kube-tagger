package pvc

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"

	"testing"
)

type TestHandler struct {
	mock.Mock
}

func (t TestHandler) Handle(volId string, tags map[string]string) {
	t.Called(volId, tags)
}

var vol = &v1.PersistentVolume{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-vol",
	},
	Spec: v1.PersistentVolumeSpec{
		PersistentVolumeSource: v1.PersistentVolumeSource{
			AWSElasticBlockStore: &v1.AWSElasticBlockStoreVolumeSource{
				VolumeID: "aws://eu-central-1c/vol-06c8c738cdfc1703c",
			},
		},
	},
}

var pvc = &v1.PersistentVolumeClaim{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-pvc",
		Annotations: map[string]string{
			"volume.beta.kubernetes.io/storage-provisioner": "kubernetes.io/aws-ebs",
		},
	},
	Spec: v1.PersistentVolumeClaimSpec{
		VolumeName: "test-vol",
	},
}

var pvcCustomTags = &v1.PersistentVolumeClaim{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-pvc",
		Annotations: map[string]string{
			"volume.beta.kubernetes.io/storage-provisioner": "kubernetes.io/aws-ebs",
			TaglistAnnotation: "second-tag=second-value,third-tag=third-value",
		},
	},
	Spec: v1.PersistentVolumeClaimSpec{
		VolumeName: "test-vol",
	},
}

func TestHandlerShouldUseDefaultTags(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	clientset := fake.NewSimpleClientset(pvc, vol)

	tHandle := &TestHandler{}
	defaultTags := map[string]string{}
	defaultTags["first-tag"] = "first-value"

	handlerCaller := HandlerCaller{
		client:      clientset,
		handler:     tHandle,
		defaultTags: defaultTags,
	}

	tHandle.On("Handle", mock.MatchedBy(func(volId string) bool {
		if volId != "vol-06c8c738cdfc1703c" {
			t.Errorf("Expected volId to be 'vol-06c8c738cdfc1703c' but was '%s'", volId)
		}
		return true
	}), mock.MatchedBy(func(tags map[string]string) bool {
		if _, ok := tags["first-tag"]; !ok {
			t.Errorf("Expected tag 'first-tag' not found")
		}
		value := tags["first-tag"]
		if value != "first-value" {
			t.Errorf("Expected tag value to be 'first-value', but was '%s'", value)
		}

		return true
	}))
	handlerCaller.handle(pvc)
}

func TestHandlerShouldUseCustomTags(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	clientset := fake.NewSimpleClientset(pvcCustomTags, vol)

	tHandle := &TestHandler{}
	defaultTags := map[string]string{}

	handlerCaller := HandlerCaller{
		client:      clientset,
		handler:     tHandle,
		defaultTags: defaultTags,
	}

	tHandle.On("Handle", mock.Anything, mock.MatchedBy(func(tags map[string]string) bool {
		t1 := tags["second-tag"]
		t2 := tags["third-tag"]

		if t1 != "second-value" {
			t.Errorf("Expected tag 'second-tag' to have value 'second-value', but was '%s'", t1)
		}
		if t2 != "third-value" {
			t.Errorf("Expected tag 'third-tag' to have value 'third-value', but was '%s'", t2)
		}

		return true
	}))
	handlerCaller.handle(pvcCustomTags)
}

func TestHandlerShouldMergeTags(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	clientset := fake.NewSimpleClientset(pvcCustomTags, vol)

	tHandle := &TestHandler{}
	defaultTags := map[string]string{}
	defaultTags["first-tag"] = "first-value"
	defaultTags["second-tag"] = "overwritten"

	handlerCaller := HandlerCaller{
		client:      clientset,
		handler:     tHandle,
		defaultTags: defaultTags,
	}

	tHandle.On("Handle", mock.Anything, mock.MatchedBy(func(tags map[string]string) bool {
		t1 := tags["first-tag"]
		t2 := tags["second-tag"]
		t3 := tags["third-tag"]

		if t1 != "first-value" {
			t.Errorf("Expected tag 'first-tag' to have value 'first-value', but was '%s'", t1)
		}
		if t2 != "second-value" {
			t.Errorf("Expected tag 'second-tag' to have value 'second-value', but was '%s'", t2)
		}
		if t3 != "third-value" {
			t.Errorf("Expected tag 'third-tag' to have value 'third-value', but was '%s'", t3)
		}

		return true
	}))
	handlerCaller.handle(pvcCustomTags)
}
