package helm

import (
	"bytes"
	"errors"
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type instanceUnit struct{}

var _ = check.Suite(&instanceUnit{})

func (iu *instanceUnit) Test_makeCmd(c *check.C) {
	err, arg := Helm.makeCmd("ls", []string{"--all"}, []string{})
	c.Assert(err, check.IsNil)
	c.Assert(arg[0], check.Equals, "ls")
	c.Assert(arg[1], check.Equals, "--all")
}

func (iu *instanceUnit) Test_execPipeLine(c *check.C) {
	err, result := Helm.executor.execPipeLine("echo", []string{"true"})
	c.Assert(err, check.IsNil)
	c.Assert(result.String()[:4], check.Matches, "true")
}

func (iu *instanceUnit) Test_MakeValueFile(c *check.C) {
	_, err := Helm.MakeValueFile([]byte("test value"))
	c.Assert(err, check.IsNil)
}

type mockShell struct {
	Shell
}

func (sh *mockShell) execPipeLine(cmd string, args []string) (error, *bytes.Buffer) {
	end := len(args) - 1
	result_1 := args[end]
	result_2 := cmd
	for _, arg := range args[:end] {
		result_2 = result_2 + " " + arg
	}
	if result_1 != result_2 {
		return errors.New("Not equal"), nil
	}
	return nil, bytes.NewBufferString("test")
}

func (iu *instanceUnit) Test_Detele(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	err := Helm.Detele([]string{"test"}, []string{"helm delete test"})
	c.Assert(err, check.IsNil)
}

func (iu *instanceUnit) Test_Rollback(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	err := Helm.Rollback([]string{"test"}, []string{"1.1", "--recreate-pods", "helm rollback test 1.1 --recreate-pods"})
	c.Assert(err, check.IsNil)
}

func (iu *instanceUnit) Test_UpdateApplications(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	err := Helm.UpdateApplications([]string{"testchart"}, []string{"--name", "test",
		"--version", "1.1",
		"--link mysql=testmysql.mysql",
		"helm upgrade testchart --name test --version 1.1 --link mysql=testmysql.mysql"})
	c.Assert(err, check.IsNil)
}

func (iu *instanceUnit) Test_UpdateRepo(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	err := Helm.UpdateRepo([]string{"update", "helm repo update"})
	c.Assert(err, check.IsNil)
}

func (iu *instanceUnit) Test_DeplyApplications(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	err := Helm.DeplyApplications([]string{"testchart"}, []string{"--name", "test",
		"--version", "1.1",
		"--link mysql=testmysql.mysql",
		"helm install testchart --name test --version 1.1 --link mysql=testmysql.mysql"})
	c.Assert(err, check.IsNil)
}

func (iu *instanceUnit) Test_StatusApplications(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	output, err := Helm.StatusApplications([]string{"testchart"}, []string{
		"--namespace", "test",
		"--reverse", "--all",
		"helm status testchart --namespace test --reverse --all"})
	c.Assert(err, check.IsNil)
	c.Assert(output, check.Equals, "test")
}

func (iu *instanceUnit) Test_ListApplications(c *check.C) {
	back := Helm.executor
	defer func() {
		Helm.executor = back
	}()
	Helm.executor = &mockShell{}
	output, err := Helm.ListApplications([]string{"test"}, []string{
		"--revision", "--output", "json",
		"helm list test --revision --output json"})
	c.Assert(err, check.IsNil)
	c.Assert(output.String(), check.Equals, "test")
}
