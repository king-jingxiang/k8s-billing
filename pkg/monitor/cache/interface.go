package cache

import (
	"github.com/ruanxingbaozi/k8s-billing/pkg/monitor/api"
)

// Cache collects pods/nodes/queues information
// and provides information snapshot
type Cache interface {
	// Run start informer
	Run(stopCh <-chan struct{})

	// Snapshot deep copy overall cache information into snapshot
	Snapshot() *api.ClusterInfo
}
