package pvc

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

type PVCController struct {
	client     kubernetes.Interface
	Controller cache.Controller
}

func NewPVCController(client kubernetes.Interface, handler IHandler, defaultTags map[string]string) (*PVCController, error) {
	pvcListWatcher := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"persistentVolumeClaims",
		v1.NamespaceAll,
		fields.Everything(),
	)

	h := HandlerCaller{
		client:      client,
		handler:     handler,
		defaultTags: defaultTags,
	}

	_, controller := cache.NewInformer(pvcListWatcher,
		&v1.PersistentVolumeClaim{},
		60*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				h.handle(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				h.handle(newObj)
			},
		},
	)

	return &PVCController{
		client:     client,
		Controller: controller,
	}, nil
}
