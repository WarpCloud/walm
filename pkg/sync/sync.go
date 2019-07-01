package sync

type Sync interface {
	Start(stopCh <-chan struct{})
}
