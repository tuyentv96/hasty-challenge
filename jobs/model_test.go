package jobs

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/utils"
)

func newTestObjectId() int {
	return int(gofakeit.Uint32())
}

func TestModelToJSON(t *testing.T) {
	job := Job{
		Id:        1,
		ObjectId:  99093383,
		Status:    JobStatusCreated,
		StartTime: utils.TimeToPtr(time.Date(2020, 02, 01, 03, 04, 05, 0, time.UTC)),
		EndTime:   utils.TimeToPtr(time.Date(2020, 03, 01, 03, 04, 05, 0, time.UTC)),
		Message:   "test message",
		CreatedAt: time.Date(2019, 04, 01, 03, 04, 05, 0, time.UTC),
	}

	want := []byte(`{"id":1,"object_id":99093383,"status":"created","start_time":"2020-02-01T03:04:05Z","end_time":"2020-03-01T03:04:05Z","message":"test message","created_at":"2019-04-01T03:04:05Z"}`)
	assert.Equal(t, want, job.ToJSON())
}

func TestModelJobFromJSON(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		job := Job{
			Id:        1,
			ObjectId:  99093383,
			Status:    JobStatusCreated,
			StartTime: utils.TimeToPtr(time.Date(2020, 02, 01, 03, 04, 05, 0, time.UTC)),
			EndTime:   utils.TimeToPtr(time.Date(2020, 03, 01, 03, 04, 05, 0, time.UTC)),
			Message:   "test message",
			CreatedAt: time.Date(2019, 04, 01, 03, 04, 05, 0, time.UTC),
		}

		js := []byte(`{"id":1,"object_id":99093383,"status":"created","start_time":"2020-02-01T03:04:05Z","end_time":"2020-03-01T03:04:05Z","message":"test message","created_at":"2019-04-01T03:04:05Z"}`)
		actual, err := JobFromJSON(js)
		require.NoError(t, err)

		assert.Equal(t, job, actual)
	})

	t.Run("invalid format", func(t *testing.T) {
		js := []byte(`{"id":"1"}`)
		_, err := JobFromJSON(js)
		assert.NotNil(t, err)
	})
}
