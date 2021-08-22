package jobs

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	JobStatusCreated JobStatus = "created"
	JobStatusRunning JobStatus = "running"
	JobStatusSuccess JobStatus = "success"
	JobStatusFailed  JobStatus = "failed"
)

type Job struct {
	tableName struct{} `pg:"jobs,discard_unknown_columns"`

	Id        int        `json:"id" pg:"id"`
	ObjectId  int        `json:"object_id" pg:"object_id"`
	Status    JobStatus  `json:"status" pg:"status"`
	StartTime *time.Time `json:"start_time" pg:"start_time"`
	EndTime   *time.Time `json:"end_time" pg:"end_time"`
	Message   string     `json:"message" pg:"message"`
	CreatedAt time.Time  `json:"created_at" pg:"created_at"`
}

func (j Job) ToJSON() []byte {
	buf, _ := json.Marshal(j)
	return buf
}

func JobFromJSON(js []byte) (Job, error) {
	job := Job{}
	err := json.Unmarshal(js, &job)
	if err != nil {
		return Job{}, err
	}

	return job, err
}

type JobPayload struct {
	ObjectId int `json:"object_id""`
}
