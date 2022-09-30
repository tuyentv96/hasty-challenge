package jobs

import (
	"context"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/benbjohnson/clock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

const (
	MinSleepTime = 15
	MaxSleepTime = 40
)

type Consumer struct {
	cfg           config.Config
	svc           Service
	logger        *logrus.Entry
	clock         clock.Clock
	random        utils.Random
	transactioner utils.Transactioner
}

func NewConsumer(cfg config.Config, logger *logrus.Entry, svc Service, clock clock.Clock, random utils.Random, transactioner utils.Transactioner) *Consumer {
	return &Consumer{
		cfg:           cfg,
		svc:           svc,
		logger:        logger,
		clock:         clock,
		random:        random,
		transactioner: transactioner,
	}
}

func (c *Consumer) Consume(delivery rmq.Delivery) {
	ctx := context.Background()
	var err error

	defer func() {
		if err == nil {
			if err := delivery.Ack(); err != nil {
				c.logger.WithError(err).Errorf("failed to ack job: %s", delivery.Payload())
			}
		} else {
			if err := delivery.Push(); err != nil {
				c.logger.WithError(err).Errorf("failed to push job: %s", delivery.Payload())
			}
		}
	}()

	var job Job
	job, err = JobFromJSON([]byte(delivery.Payload()))
	if err != nil {
		c.logger.WithError(err).Errorf("failed to parse job: %s", delivery.Payload())
		return
	}

	err = c.DoJob(ctx, job)
}

func (c *Consumer) DoJob(ctx context.Context, job Job) error {
	// Wrap DoJob function with a transaction
	return c.transactioner.RunWithTransaction(ctx, func(ctx context.Context) error {
		return c.doJob(ctx, job)
	})
}

func (c *Consumer) doJob(ctx context.Context, job Job) (err error) {
	// Try to claim job
	err = c.svc.ClaimJob(ctx, job)
	if err != nil {
		// Job was claimed by another worker, just ignore
		if errors.Is(err, ErrJobWasClaimed) {
			c.logger.WithField("jobId", job.Id).Errorf("Job was claimed by another worker")
			return nil
		}

		return err
	}

	var isJobTimeout bool

	defer func() {
		if isJobTimeout {
			_, err = c.svc.SetJobFailed(ctx, job, ErrJobExceedTimeout.Error())
			if err != nil {
				err = errors.Wrap(err, "failed to set job failed")
			} else {
				c.logger.WithField("jobId", job.Id).Error("Job was exceed timeout")
			}
		} else {
			_, err = c.svc.SetJobSuccess(ctx, job)
			if err != nil {
				err = errors.Wrap(err, "failed to set job success")
			} else {
				c.logger.WithField("jobId", job.Id).Infof("Job ran successfully")
			}
		}
	}()

	sleepTime := c.random.Rand(MinSleepTime, MaxSleepTime)
	c.logger.WithField("jobId", job.Id).Infof("Job will run in %d seconds", sleepTime)

	select {
	case <-c.clock.After(time.Duration(sleepTime) * time.Second):
		break
	case <-c.clock.After(time.Duration(c.cfg.JobConfig.TimeoutInSeconds) * time.Second):
		isJobTimeout = true
	}

	return err
}
