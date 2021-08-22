package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/tuyentv96/hasty-challenge/utils"
)

type Store interface {
	SaveJob(ctx context.Context, job Job) (Job, error)
	UpdateJobOptimistically(ctx context.Context, job Job, currentStatus JobStatus) error
	GetJobByID(ctx context.Context, jobId int) (Job, error)
	GetJobByObjectId(ctx context.Context, objectId int, createdAt time.Time) (Job, error)
}

type StoreImpl struct {
	db orm.DB
}

func (j StoreImpl) GetDB(ctx context.Context) orm.DB {
	return utils.TransactionFromContext(ctx, j.db)
}

func NewJobStore(db orm.DB) *StoreImpl {
	return &StoreImpl{
		db: db,
	}
}

func (j StoreImpl) SaveJob(ctx context.Context, job Job) (Job, error) {
	err := j.GetDB(ctx).Insert(&job)
	if err != nil {
		return Job{}, err
	}

	return job, nil
}

func (j StoreImpl) UpdateJobOptimistically(ctx context.Context, job Job, currentStatus JobStatus) error {
	result, err := j.GetDB(ctx).Model(&job).
		Set("status = ?", job.Status).
		Set("start_time = ?", job.StartTime).
		Set("end_time = ?", job.EndTime).
		Set("message = ?", job.Message).
		Where("id = ?", job.Id).
		Where("status = ?", currentStatus).
		Update()
	if err != nil {
		return err
	}

	if count := result.RowsAffected(); count == 0 {
		return ErrNoRowUpdated
	}

	return nil
}

func (j StoreImpl) GetJobByID(ctx context.Context, jobId int) (Job, error) {
	var result Job

	if err := j.GetDB(ctx).Model(&result).
		Where("id = ?", jobId).
		Select(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return Job{}, ErrJobNotFound
		}

		return Job{}, err
	}

	return result, nil
}

func (j StoreImpl) GetJobByObjectId(ctx context.Context, objectId int, createdAt time.Time) (Job, error) {
	var result Job

	if err := j.GetDB(ctx).Model(&result).
		Where("object_id = ?", objectId).
		Where("created_at >= ?", createdAt).
		Limit(1).
		Select(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return Job{}, ErrJobNotFound
		}

		return Job{}, err
	}

	return result, nil
}
