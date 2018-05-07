package helm

import (
	"bytes"
	"errors"
	"log"
	"strings"

	"gopkg.in/pipe.v2"
)

type Interface struct {
	cmd string
}

var Helm *Interface

func init() {
	Helm = &Interface{cmd: "helm"}
}

func (inst *Interface) makeCmd(args, flags []string) (error, string) {
	if len(args) == 0 && len(flags) == 0 {
		return errors.New("no args and no flags"), ""
	}
	cmd := inst.cmd
	cmd += strings.Join(args, " ")
	cmd += strings.Join(flags, " ")
	return nil, cmd
}

func (inst *Interface) Detele(args, flags []string) error {
	if err, cmd := inst.makeCmd(args, flags); err != nil {
		return err
	} else {
		b := &bytes.Buffer{}
		p := pipe.Line(
			pipe.Exec(cmd),
			pipe.Write(b),
		)
		err = pipe.Run(p)
		if err != nil {
			log.Printf("%v\n", err)
		}
	}

	return nil
}
func (inst *Interface) Rollback(args, flags []string) error {
	return nil
}
func (inst *Interface) DeplyApplications(args, flags []string) error {
	return nil
}
func (inst *Interface) UpdateApplications(args, flags []string) error {
	return nil
}
func (inst *Interface) GetApplicationsStatus(args, flags []string) (string, error) {
	return "", nil
}
func (inst *Interface) FindApplicationsStatus(args, flags []string) ([]Status, error) {
	return nil, nil
}
