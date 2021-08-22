package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestStoreSaveJob(t *testing.T) {
	job := Job{
		ObjectId: newTestObjectId(),
		Status:   JobStatusCreated,
	}

	cases := []struct {
		name string
		job  Job
		err  error
	}{
		{
			name: "create new job",
			job:  job,
			err:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testStore.SaveJob(context.Background(), tc.job)
			assert.Equal(t, tc.err, err)
		})

	}
}

func TestStoreGetJob(t *testing.T) {
	ctx := context.Background()

	job := Job{
		ObjectId: newTestObjectId(),
		Status:   JobStatusCreated,
	}
	job, err := testStore.SaveJob(ctx, job)
	require.NoError(t, err)

	cases := []struct {
		name  string
		jobId int
		err   error
	}{
		{
			name:  "get existede job",
			jobId: job.Id,
			err:   nil,
		},
		{
			name:  "get non existd job",
			jobId: 99999,
			err:   ErrJobNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testStore.GetJobByID(context.Background(), tc.jobId)
			assert.Equal(t, tc.err, err)
		})

	}
}

func TestStoreGetJobByObjectId(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	job := Job{
		ObjectId:  newTestObjectId(),
		Status:    JobStatusCreated,
		CreatedAt: now.Add(5 * time.Minute),
	}
	job, err := testStore.SaveJob(ctx, job)
	require.NoError(t, err)

	cases := []struct {
		name      string
		objectId  int
		createdAt time.Time
		err       error
	}{
		{
			name:     "get exist object_id",
			objectId: newTestObjectId(),
			err:      ErrJobNotFound,
		},
		{
			name:      "get job successfully",
			objectId:  job.ObjectId,
			createdAt: now,
			err:       nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := testStore.GetJobByObjectId(context.Background(), tc.objectId, tc.createdAt)
			assert.Equal(t, tc.err, err)
		})
	}
}

func TestStoreUpdateJob(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name          string
		job           func() Job
		currentStatus JobStatus
		err           error
	}{
		{
			name:          "update no exist job",
			currentStatus: JobStatusCreated,
			job: func() Job {
				return Job{}
			},
			err: ErrNoRowUpdated,
		},
		{
			name:          "update successfully",
			currentStatus: JobStatusCreated,
			job: func() Job {
				job := Job{
					ObjectId: newTestObjectId(),
					Status:   JobStatusCreated,
				}

				job, err := testStore.SaveJob(ctx, job)
				require.NoError(t, err)
				job.Status = JobStatusRunning
				return job
			},
			err: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := testStore.UpdateJobOptimistically(ctx, tc.job(), tc.currentStatus)
			assert.Equal(t, tc.err, err)
		})
	}
}
