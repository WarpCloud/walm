package task

type TaskSig struct {
	UUID        string              `json:"uuid" description:"task uuid"`
	Name        string              `json:"name" description:"task name"`
	Arg         string              `json:"arg" description:"task arg"`
	TimeoutSec  int64               `json:"timeout_sec" description:"task timeout(sec)"`
}
