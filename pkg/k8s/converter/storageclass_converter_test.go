package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)
func TestConvertStorageClassFromK8s(t *testing.T) {
	testAllowVolumeExpansion := false
	testPersistentVolumeReclaimPolicy := corev1.PersistentVolumeReclaimPolicy("Retain")
	testVolumeBindingMode := v1.VolumeBindingMode("Immediate")
	tests := []struct{
		oriStorageClass *v1.StorageClass
		storageClass    *k8s.StorageClass
		err             error
	}{
		{
			oriStorageClass: &v1.StorageClass{
				TypeMeta:             metav1.TypeMeta{
					Kind: "StorageClass",
				},
				ObjectMeta:           metav1.ObjectMeta{
					Name: "standard",
					Namespace: "test-namespace",
				},
				Provisioner:          "kubernetes.io/aws-ebs",
				ReclaimPolicy:        &testPersistentVolumeReclaimPolicy,
				AllowVolumeExpansion: &testAllowVolumeExpansion,
				VolumeBindingMode:    &testVolumeBindingMode,
			},
			storageClass: &k8s.StorageClass{
				Meta:                 k8s.Meta{
					Name: "standard",
					Namespace: "test-namespace",
					Kind: "StorageClass",
					State: k8s.State{
						Status:  "Ready",
						Reason:  "",
						Message: "",
					},
				},
				Provisioner:          "kubernetes.io/aws-ebs",
				ReclaimPolicy:        "Retain",
				AllowVolumeExpansion: false,
				VolumeBindingMode:    "Immediate",
			},
			err: nil,
		},
		{
			oriStorageClass: nil,
			storageClass: nil,
			err:         nil,
		},
	}

	for _, test := range tests {
		storageClass, err := ConvertStorageClassFromK8s(test.oriStorageClass)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.storageClass, storageClass)
	}
}