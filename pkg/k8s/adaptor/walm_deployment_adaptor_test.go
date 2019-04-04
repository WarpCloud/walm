package adaptor

import (
	"testing"
	"fmt"
)

func TestParsePod(t *testing.T) {
	walmPods := []*WalmPod{buildWalmPod("test1", "Pending"), buildWalmPod("test2", "Terminating")}
	buildWalmPod("test1", "Terminating")

	allPodsTerminating, unknownPod, pendingPod, runningPod := parsePods(walmPods)

	fmt.Println(allPodsTerminating, unknownPod, pendingPod, runningPod)
}

func buildWalmPod(name string, state string) *WalmPod {
	return &WalmPod{
		WalmMeta: WalmMeta{Name: name, State: WalmState{Status: state}},
	}
}
