package main

import (
	"testing"
)

func Test_Lint_Chart(t *testing.T) {
	lint := lintOptions{
		chartPath: "/home/bianyu/codes/application-helmcharts/transwarp-jsonnetcharts/demo/1.0",
	}
	lint.run()
}
