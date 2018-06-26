package helm

import (
	"bytes"
	"errors"

	"io/ioutil"

	. "walm/pkg/util/log"

	"gopkg.in/pipe.v2"
)

type Cmd interface {
	execPipeLine(cmd string, args []string) (error, *bytes.Buffer)
}

type Shell struct{}

type Interface struct {
	cmd      string
	path     string
	executor Cmd
}

var Helm *Interface

func init() {
	Helm = &Interface{cmd: "helm", executor: &Shell{}}
}

func (inst *Interface) makeCmd(subcmd string, args, flags []string) (error, []string) {
	if len(args) == 0 && len(flags) == 0 {
		return errors.New("no args and no flags"), nil
	}
	argarray := []string{}
	argarray = append(argarray, subcmd)
	for _, arg := range args {
		argarray = append(argarray, arg)
	}
	for _, flag := range flags {
		argarray = append(argarray, flag)
	}

	return nil, argarray
}

func (sh *Shell) execPipeLine(cmd string, args []string) (error, *bytes.Buffer) {
	Log.Debugf("beging to exec cmd: %s", cmd)
	defer Log.Debugf("end to exec cmd: %s", cmd)
	b := &bytes.Buffer{}
	p := pipe.Line(
		pipe.Exec(cmd, args...),
		pipe.Write(b),
	)
	err := pipe.Run(p)
	return err, b
}

// MakeValueFile creates a temporary file in TempDir (see os.TempDir)
// and writes values to the file and resturn its name. It is the caller's responsibility
// to remove the file returned if necessary.
func (inst *Interface) MakeValueFile(data []byte) (string, error) {
	tmpFile, err := ioutil.TempFile("", "tmp-values-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err = tmpFile.Write(data); err != nil {
		return tmpFile.Name(), err
	}
	tmpFile.Sync()
	return tmpFile.Name(), nil
}

func (inst *Interface) Detele(args, flags []string) error {
	if err, arg := inst.makeCmd("delete", args, flags); err != nil {
		return err
	} else {
		err, _ = inst.executor.execPipeLine(inst.cmd, arg)
		return err
	}
}
func (inst *Interface) Rollback(args, flags []string) error {
	if err, arg := inst.makeCmd("rollback", args, flags); err != nil {
		return err
	} else {
		err, _ = inst.executor.execPipeLine(inst.cmd, arg)
		return err
	}
}

func (inst *Interface) UpdateRepo(args []string) error {
	if err, arg := inst.makeCmd("repo", args, []string{}); err != nil {
		return err
	} else {
		err, _ = inst.executor.execPipeLine(inst.cmd, arg)
		return err
	}
}

func (inst *Interface) DeplyApplications(args, flags []string) error {
	//update repo before install chart
	/*
		if err := inst.UpdateRepo([]string{"update"}); err != nil {
			return err
		}
	*/

	if err, arg := inst.makeCmd("install", args, flags); err != nil {
		return err
	} else {
		err, _ = inst.executor.execPipeLine(inst.cmd, arg)
		return err
	}
}
func (inst *Interface) UpdateApplications(args, flags []string) error {
	if err, arg := inst.makeCmd("upgrade", args, flags); err != nil {
		return err
	} else {
		err, _ = inst.executor.execPipeLine(inst.cmd, arg)
		return err
	}
}

func (inst *Interface) StatusApplications(args, flags []string) (string, error) {
	if err, arg := inst.makeCmd("status", args, flags); err != nil {
		return "", err
	} else {
		var b *bytes.Buffer
		err, b = inst.executor.execPipeLine(inst.cmd, arg)
		return b.String(), err
	}
}

func (inst *Interface) ListApplications(args, flags []string) (*bytes.Buffer, error) {

	if err, arg := inst.makeCmd("list", args, flags); err != nil {
		return &bytes.Buffer{}, err
	} else {
		var b *bytes.Buffer
		err, b = inst.executor.execPipeLine(inst.cmd, arg)
		return b, err
	}
}
