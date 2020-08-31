package pvc

import v1 "k8s.io/api/core/v1"

func isEBSVolume(volume *v1.PersistentVolumeClaim) bool {
	for k, v := range volume.Annotations {
		if k == "volume.beta.kubernetes.io/storage-provisioner" && v == "kubernetes.io/aws-ebs" {
			return true
		}
	}
	return false
}
