package helm

import (
	"testing"
)

func Test_ChartImage(t *testing.T) {
	chart, err := GetDefaultChartImageClient().GetChart("172.16.1.41:5000/cy-charts/hdfs:1")
	if err != nil {
		t.Error(err.Error())
	} else {
		if chart.Name() != "hdfs" {
			t.Fail()
		}
	}
}
