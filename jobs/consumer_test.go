package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

func initTestConsumer(svc Service, clock clock.Clock, random utils.Random) *Consumer {
	cfg := config.Config{}
	return NewConsumer(cfg, testLogger, svc, clock, random, testTransaction)
}

func TestConsumerDoJob(t *testing.T) {
	t.Run("do job successfully", func(t *testing.T) {
		ctx := context.Background()
		clock := clock.NewMock()
		random := utils.NewMockRandomImpl()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		consumer := initTestConsumer(svc, clock, random)
		job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		job, err = svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)

		consumer.cfg.JobConfig.TimeoutInSeconds = 30
		random.SetVal(25)

		wait := make(chan bool)
		go func() {
			err = consumer.DoJob(ctx, job)
			close(wait)
		}()

		time.Sleep(time.Second)
		clock.Add(25 * time.Second)
		<-wait
		require.NoError(t, err)

		job, err = svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusSuccess, job.Status)
	})

	t.Run("job exceed timeout", func(t *testing.T) {
		ctx := context.Background()
		clock := clock.NewMock()
		random := utils.NewMockRandomImpl()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		consumer := initTestConsumer(svc, clock, random)
		job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		job, err = svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)

		consumer.cfg.JobConfig.TimeoutInSeconds = 30
		random.SetVal(35)

		wait := make(chan bool)
		go func() {
			err = consumer.DoJob(ctx, job)
			close(wait)
		}()

		time.Sleep(time.Second)
		clock.Add(30 * time.Second)
		<-wait
		require.Error(t, ErrJobExceedTimeout, err)

		job, err = svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusFailed, job.Status)
	})

	t.Run("job was claimed", func(t *testing.T) {
		ctx := context.Background()
		clock := clock.NewMock()
		random := utils.NewMockRandomImpl()
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		consumer := initTestConsumer(svc, clock, random)
		job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		err = svc.ClaimJob(ctx, job)
		require.NoError(t, err)

		consumer.cfg.JobConfig.TimeoutInSeconds = 30
		random.SetVal(35)

		err = consumer.DoJob(ctx, job)
		require.Nil(t, err)

		job, err = svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusRunning, job.Status)
	})
}
