package jobs

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

func initTestWorker(t *testing.T, cfg config.Config, svc Service, queueName string, clock clock.Clock, random utils.Random) *WorkerImpl {
	queue := initTestQueue(t, queueName)
	return NewWorker(cfg, testLogger, svc, testRmqConnection, queue, clock, random, testTransaction)
}

func TestWorkerStartAndStop(t *testing.T) {
	clock := initTestClock()
	random := utils.NewMockRandomImpl()
	cfg := config.Config{
		JobConfig: config.JobConfig{
			TimeoutInSeconds: 30,
			JobPrefetch:      5,
		},
		RedisConfig: config.RedisConfig{
			RedisPollIntervalMs: 100,
		},
	}
	queueName := gofakeit.UUID()
	svc := initTestService(t, queueName, clock)
	worker := initTestWorker(t, cfg, svc, queueName, clock, random)

	go func() {
		<-time.After(100 * time.Millisecond)
		worker.Stop()
	}()

	err := worker.Start()
	require.NoError(t, err)
}
