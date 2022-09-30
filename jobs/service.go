package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/adjust/rmq/v5"

	"github.com/tuyentv96/hasty-challenge/utils"
)

const TimeWindowInMinutes = 5

type Service interface {
	SaveJob(ctx context.Context, payload JobPayload) (Job, error)
	ClaimJob(ctx context.Context, job Job) error
	GetJobByID(ctx context.Context, jobId int) (Job, error)
	PublishJob(ctx context.Context, job Job) error
	SetJobFailed(ctx context.Context, job Job, message string) (Job, error)
	SetJobSuccess(ctx context.Context, job Job) (Job, error)
}

type ServiceImpl struct {
	store Store
	queue rmq.Queue
	clock clock.Clock
}

func NewService(store Store, queue rmq.Queue, clock clock.Clock) *ServiceImpl {
	return &ServiceImpl{
		store: store,
		queue: queue,
		clock: clock,
	}
}

func (s *ServiceImpl) SaveJob(ctx context.Context, payload JobPayload) (Job, error) {
	job := Job{
		ObjectId: payload.ObjectId,
	}

	timeWindow := s.clock.Now().Add(-time.Duration(TimeWindowInMinutes) * time.Minute)
	existJob, err := s.store.GetJobByObjectId(ctx, job.ObjectId, timeWindow)
	if err != nil && !errors.Is(err, ErrJobNotFound) {
		return Job{}, err
	}

	if err == nil {
		return existJob, nil
	}

	job.CreatedAt = s.clock.Now().UTC()
	job.Status = JobStatusCreated
	job, err = s.store.SaveJob(ctx, job)
	if err != nil {
		return Job{}, nil
	}

	if err := s.PublishJob(ctx, job); err != nil {
		return Job{}, err
	}

	return job, nil
}

func (s *ServiceImpl) GetJobByID(ctx context.Context, jobId int) (Job, error) {
	return s.store.GetJobByID(ctx, jobId)
}

func (s *ServiceImpl) ClaimJob(ctx context.Context, job Job) error {
	job.Status = JobStatusRunning
	job.StartTime = utils.TimeToPtr(s.clock.Now())

	if err := s.store.UpdateJobOptimistically(ctx, job, JobStatusCreated); err != nil {
		if errors.Is(err, ErrNoRowUpdated) {
			return ErrJobWasClaimed
		}

		return err
	}

	return nil
}

func (s *ServiceImpl) PublishJob(ctx context.Context, job Job) error {
	return s.queue.PublishBytes(job.ToJSON())
}

func (s *ServiceImpl) SetJobSuccess(ctx context.Context, job Job) (Job, error) {
	job.Status = JobStatusSuccess
	job.EndTime = utils.TimeToPtr(s.clock.Now())
	if err := s.store.UpdateJobOptimistically(ctx, job, JobStatusRunning); err != nil {
		if errors.Is(err, ErrNoRowUpdated) {
			return Job{}, ErrJobWasNotClaimed
		}

		return Job{}, err
	}

	return job, nil
}

func (s *ServiceImpl) SetJobFailed(ctx context.Context, job Job, message string) (Job, error) {
	job.Status = JobStatusFailed
	job.EndTime = utils.TimeToPtr(s.clock.Now())
	job.Message = message
	if err := s.store.UpdateJobOptimistically(ctx, job, JobStatusRunning); err != nil {
		if errors.Is(err, ErrNoRowUpdated) {
			return Job{}, ErrJobWasNotClaimed
		}

		return Job{}, err
	}

	return job, nil
}
