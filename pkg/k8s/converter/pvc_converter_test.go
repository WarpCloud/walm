package converter

import (
	"WarpCloud/walm/pkg/models/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConvertPvcFromK8s(t *testing.T) {
	testStorageName := "test-storage-class"
	testPersistentVolumeMode := corev1.PersistentVolumeMode("Filesystems")
	tests := []struct {
		oriPersistentVolumeClaim *corev1.PersistentVolumeClaim
		pvc                      *k8s.PersistentVolumeClaim
		err                      error
	}{
		{
			oriPersistentVolumeClaim: &corev1.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind: "PersistentVolumeClaim",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-pvc",
					Labels: map[string]string{"test1": "test1"},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						"ReadWriteOnce",
					},
					VolumeMode:       &testPersistentVolumeMode,
					VolumeName:       "test-volume-class",
					StorageClassName: &testStorageName,
				},
				Status: corev1.PersistentVolumeClaimStatus{
					Capacity: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("8Gi"),
					},
				},
			},
			pvc: &k8s.PersistentVolumeClaim{
				Meta: k8s.Meta{
					Name: "test-pvc",
					Kind: "PersistentVolumeClaim",
				},
				Labels:       map[string]string{"test1": "test1"},
				StorageClass: "test-storage-class",
				VolumeName:   "test-volume-class",
				Capacity:     "8Gi",
				AccessModes:  []string{"ReadWriteOnce"},
				VolumeMode:   "Filesystems",
			},
		},
		{
			oriPersistentVolumeClaim: nil,
			pvc:                      nil,
			err:                      nil,
		},
	}

	for _, test := range tests {
		pvc, err := ConvertPvcFromK8s(test.oriPersistentVolumeClaim)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.pvc, pvc)
	}
}
