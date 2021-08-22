package jobs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

func TestRunJobSuccessfully(t *testing.T) {
	ctx := context.Background()
	now := utils.TimeNow()
	t.Run("run job successfully", func(t *testing.T) {
		objectId := newTestObjectId()
		clock := initTestClock()
		random := utils.NewMockRandomImpl()
		cfg := config.Config{
			JobConfig: config.JobConfig{
				TimeoutInSeconds: 40,
			},
			RedisConfig: config.RedisConfig{
				RedisPrefetch:       1,
				RedisPollIntervalMs: 100,
			},
		}
		queueName := fmt.Sprintf("run_job_%s", gofakeit.UUID())
		svc := initTestService(t, queueName, clock)

		worker := initTestWorker(t, cfg, svc, queueName, clock, random)
		go worker.Start()

		handler := initTestHandler(cfg, svc)

		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(fmt.Sprintf(`{"object_id": %d}`, objectId)),
		}

		sleepTimeInSeconds := 35
		clock.Set(now)
		random.SetVal(sleepTimeInSeconds)

		rec := tr.do(handler)
		assert.Equal(t, http.StatusCreated, rec.Code)
		job := jobFromRec(t, rec)

		time.Sleep(2 * time.Second)
		clock.Add(time.Duration(sleepTimeInSeconds) * time.Second)
		time.Sleep(2 * time.Second)

		actual, err := svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusSuccess, actual.Status)
	})

	t.Run("run job exceed timeout", func(t *testing.T) {
		clock := initTestClock()
		random := utils.NewMockRandomImpl()
		cfg := config.Config{
			JobConfig: config.JobConfig{
				TimeoutInSeconds: 25,
			},
			RedisConfig: config.RedisConfig{
				RedisPrefetch:       1,
				RedisPollIntervalMs: 100,
			},
		}
		queueName := fmt.Sprintf("job_exceed_timeout_%s", gofakeit.UUID())
		svc := initTestService(t, queueName, clock)

		worker := initTestWorker(t, cfg, svc, queueName, clock, random)
		go func() {
			err := worker.Start()
			if err != nil {
				log.Fatalln(err.Error())
			}
		}()

		objectId := newTestObjectId()
		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(fmt.Sprintf(`{"object_id": %d}`, objectId)),
		}

		sleepTimeInSeconds := 30
		clock.Set(now)
		random.SetVal(sleepTimeInSeconds)

		handler := initTestHandler(cfg, svc)
		rec := tr.do(handler)
		assert.Equal(t, http.StatusCreated, rec.Code)
		job := jobFromRec(t, rec)

		time.Sleep(2 * time.Second)
		clock.Add(time.Duration(cfg.TimeoutInSeconds) * time.Second)
		time.Sleep(2 * time.Second)

		actual, err := svc.GetJobByID(ctx, job.Id)
		require.NoError(t, err)
		assert.Equal(t, JobStatusFailed, actual.Status)
		assert.Equal(t, objectId, actual.ObjectId)
		assert.Equal(t, ErrJobExceedTimeout.Error(), actual.Message)
	})
}
