package jobs

import (
	"fmt"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/benbjohnson/clock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

const (
	QueueName = "job-queue"
)

type Worker interface {
	Start() error
	Stop()
	RunCleaner()
}

type WorkerImpl struct {
	cfg           config.Config
	svc           Service
	connection    rmq.Connection
	queue         rmq.Queue
	closed        chan bool
	logger        *logrus.Entry
	clock         clock.Clock
	random        utils.Random
	transactioner utils.Transactioner
}

func NewWorker(cfg config.Config, logger *logrus.Entry, svc Service, connection rmq.Connection, queue rmq.Queue, clock clock.Clock, random utils.Random, transactioner utils.Transactioner) *WorkerImpl {
	return &WorkerImpl{
		cfg:           cfg,
		closed:        make(chan bool),
		svc:           svc,
		connection:    connection,
		queue:         queue,
		logger:        logger.WithField("tag", "worker"),
		clock:         clock,
		random:        random,
		transactioner: transactioner,
	}
}

func (w *WorkerImpl) Start() error {
	if err := w.queue.StartConsuming(w.cfg.JobPrefetch, time.Duration(w.cfg.RedisConfig.RedisPollIntervalMs)*time.Millisecond); err != nil {
		return errors.Wrapf(err, "failed to start consuming")
	}

	for i := int64(0); i < w.cfg.JobPrefetch; i++ {
		if _, err := w.queue.AddConsumer(fmt.Sprintf("worker:%d", i), NewConsumer(w.cfg, w.logger, w.svc, w.clock, w.random, w.transactioner)); err != nil {
			return errors.Wrap(err, "failed to add consumer")
		}
	}

	w.logger.Info("Start worker successfully")
	// wait until channel is closed
	<-w.closed
	return nil
}

func (w *WorkerImpl) Stop() {
	<-w.queue.StopConsuming()
	close(w.closed)
}

// RunCleaner cleaner to make sure no unacked deliveries are stuck in the queue system.
// it will detect queue connections whose heartbeat expired and will clean up all their consumer queues by moving their unacked deliveries back to the ready list.
func (w *WorkerImpl) RunCleaner() {
	cleaner := rmq.NewCleaner(w.connection)

	for {
		select {
		case <-time.After(time.Second):
			returned, err := cleaner.Clean()
			if err != nil {
				w.logger.WithError(err).Error("[rmq] failed to clean")
				continue
			}

			if returned > 0 {
				w.logger.Infof("[rmq] cleaned %d msg", returned)
			}
		case <-w.closed:
			return
		}
	}
}
