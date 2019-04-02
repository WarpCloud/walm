package util

const (
	K8sResourceMemoryScale  int64   = 1024 * 1024
	K8sResourceStorageScale int64   = 1024 * 1024 * 1024
	K8sResourceCpuScale     float64 = 1000

	// k8s resource memory unit
	K8sResourceMemoryUnit = "Mi"

	// k8s resource storage unit
	K8sResourceStorageUnit = "Gi"
)