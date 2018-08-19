package helm

import (
	"testing"
	"fmt"
)

func Test_Repo_List(t *testing.T) {
	chartInfo, err := GetChartInfo("stable", "hdfs", "")
	fmt.Printf("%+v %v", chartInfo, err)
}
