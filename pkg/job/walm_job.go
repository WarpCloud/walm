package job

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"fmt"
)

var jobTypes map[string]Job

type WalmJobStatus string

const (
	jobStatusPending WalmJobStatus = "Pending"
	jobStatusRunning WalmJobStatus = "Running"
)

type WalmJob struct {
	JobType string        `json:"job_type" description:"job type"`
	Job     Job           `json:"job" description:"job object"`
	Id      string        `json:"id" description:"job id"`
	Status  WalmJobStatus `json:"status" description:"job status"`
}

type WalmJobAdaptor struct {
	JobType string          `json:"job_type" description:"job type"`
	Job     json.RawMessage `json:"job" description:"job object"`
	Id      string          `json:"id" description:"job id"`
	Status  WalmJobStatus   `json:"status" description:"job status"`
}

func (adaptor *WalmJobAdaptor) getJobByType() (Job, error) {
	if job, ok := jobTypes[adaptor.JobType]; ok {
		err := json.Unmarshal(adaptor.Job, job)
		if err != nil {
			logrus.Errorf("failed to unmarshal job %s : %s", string(adaptor.Job), err.Error())
			return nil , err
		}
		return job, err
	} else {
		err := fmt.Errorf("job type %s is not supported", adaptor.JobType)
		logrus.Errorf(err.Error())
		return nil, err
	}
}

func (adaptor *WalmJobAdaptor) GetWalmJob() (walmJob *WalmJob, err error) {
	walmJob = &WalmJob{
		Id:      adaptor.Id,
		JobType: adaptor.JobType,
		Status:  adaptor.Status,
	}
	walmJob.Job, err = adaptor.getJobByType()
	return
}

func (walmJob *WalmJob) Run() {
	logrus.Infof("start to run walm job %s", walmJob.Id)
	err := walmJob.Job.Do()
	if err != nil {
		//TODO retry
		logrus.Errorf("failed to run walm job %s: %s", walmJob.Id, err.Error())
	} else {
		logrus.Infof("succeed to run walm job %s", walmJob.Id)
	}
}

type Job interface {
	Do() error
}

func RegisterJobType(jobType string, job Job) {
	if jobTypes == nil {
		jobTypes = map[string]Job{}
	}
	jobTypes[jobType] = job
}
