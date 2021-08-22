package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/utils"
)

func initTestClock() *clock.Mock {
	return clock.NewMock()
}

func initTestQueue(t *testing.T, queueName string) rmq.Queue {
	queue, err := testRmqConnection.OpenQueue(queueName)
	require.NoError(t, err)
	return queue
}

func initTestService(t *testing.T, queueName string, clock clock.Clock) *ServiceImpl {
	queue := initTestQueue(t, queueName)

	return &ServiceImpl{
		store: testStore,
		queue: queue,
		clock: clock,
	}
}

func TestServiceSaveJob(t *testing.T) {
	ctx := context.Background()
	now := utils.TimeNow()

	t.Run("save new object successfully", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)
		clock.Set(now)

		payload := JobPayload{ObjectId: newTestObjectId()}
		actual, err := svc.SaveJob(ctx, payload)
		require.NoError(t, err)
		assert.NotZero(t, actual.Id)
		assert.Equal(t, payload.ObjectId, actual.ObjectId)
		assert.Equal(t, JobStatusCreated, actual.Status)
	})

	t.Run("save same object_id in five minutes, return same job", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		objectId := newTestObjectId()
		payload := JobPayload{ObjectId: objectId}

		clock.Set(now.Add(-2 * time.Minute))
		job, err := svc.SaveJob(ctx, payload)
		require.NoError(t, err)

		clock.Set(now)
		actual, err := svc.SaveJob(ctx, payload)
		require.NoError(t, err)
		assert.Equal(t, job.Id, actual.Id)
		assert.Equal(t, job.ObjectId, actual.ObjectId)
		assert.Equal(t, job.Status, actual.Status)
	})

	t.Run("save same object_id before five minutes, return new job", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		payload := JobPayload{ObjectId: newTestObjectId()}

		clock.Set(now.Add(-60 * time.Minute))
		job1, err := svc.SaveJob(ctx, payload)
		require.NoError(t, err)

		clock.Set(now)
		actual, err := svc.SaveJob(ctx, payload)
		require.NoError(t, err)
		assert.NotEqual(t, job1.Id, actual.Id)
		assert.Equal(t, payload.ObjectId, actual.ObjectId)
		assert.Equal(t, JobStatusCreated, actual.Status)
	})
}

func TestServiceGetJob(t *testing.T) {
	ctx := context.Background()

	clock := initTestClock()
	queueName := gofakeit.UUID()
	svc := initTestService(t, queueName, clock)

	job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
	require.NoError(t, err)

	cases := []struct {
		name string
		id   int
		err  error
	}{
		{
			name: "get non exist id",
			id:   int(gofakeit.Int32()),
			err:  ErrJobNotFound,
		},
		{
			name: "get job successfully",
			id:   job.Id,
			err:  nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := svc.GetJobByID(ctx, tc.id)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				assert.Equal(t, tc.id, actual.Id)
			}
		})
	}
}

func TestServiceClaimJob(t *testing.T) {
	ctx := context.Background()
	now := utils.TimeNow()

	t.Run("claim job successfully", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		clock.Set(now)
		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		err = svc.ClaimJob(ctx, job1)
		require.NoError(t, err)

		actual, err := svc.GetJobByID(ctx, job1.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusRunning, actual.Status)
		assert.Equal(t, now.Second(), actual.StartTime.Second())
	})

	t.Run("claim job two times", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		err = svc.ClaimJob(ctx, job1)
		require.NoError(t, err)

		// Try to claim job again
		err = svc.ClaimJob(ctx, job1)
		require.Equal(t, ErrJobWasClaimed, err)
	})
}

func TestServiceSetJobSuccess(t *testing.T) {
	ctx := context.Background()
	now := utils.TimeNow()

	t.Run("set job successfully", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		clock.Set(now)
		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		err = svc.ClaimJob(ctx, job1)
		require.NoError(t, err)

		job1, err = svc.GetJobByID(ctx, job1.Id)
		require.NoError(t, err)

		job1, err = svc.SetJobSuccess(ctx, job1)
		require.NoError(t, err)

		actual, err := svc.GetJobByID(ctx, job1.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusSuccess, actual.Status)
		assert.Equal(t, now.Second(), actual.EndTime.Second())
	})

	t.Run("job was not claimed", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		job1, err = svc.SetJobSuccess(ctx, job1)
		require.Error(t, err, ErrJobWasNotClaimed)
	})
}

func TestServiceSetJobFailed(t *testing.T) {
	ctx := context.Background()
	now := utils.TimeNow()

	t.Run("set job failed successfully", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)
		clock.Set(now)

		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		err = svc.ClaimJob(ctx, job1)
		require.NoError(t, err)

		job1, err = svc.GetJobByID(ctx, job1.Id)
		require.NoError(t, err)

		msg := "test message"
		job1, err = svc.SetJobFailed(ctx, job1, msg)
		require.NoError(t, err)

		actual, err := svc.GetJobByID(ctx, job1.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusFailed, actual.Status)
		assert.Equal(t, msg, actual.Message)
		assert.Equal(t, now.Second(), actual.EndTime.Second())
	})

	t.Run("job was not claimed", func(t *testing.T) {
		clock := initTestClock()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		msg := "test message"
		job1, err = svc.SetJobFailed(ctx, job1, msg)
		require.Error(t, err, ErrJobWasNotClaimed)
	})
}

func TestServicePublicJobFailed(t *testing.T) {
	ctx := context.Background()
	clock := initTestClock()
	queueName := gofakeit.UUID()
	svc := initTestService(t, queueName, clock)

	job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
	require.NoError(t, err)

	err = svc.PublishJob(ctx, job)
	require.NoError(t, err)
}
