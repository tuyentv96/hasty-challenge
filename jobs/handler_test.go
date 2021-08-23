package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/utils"
)

func initTestHandler(cfg config.Config, svc Service) *HTTPHandler {
	return NewHTTPHandler(cfg, svc)
}

func jobFromRec(t *testing.T, rec *httptest.ResponseRecorder) Job {
	var resp Job
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotZero(t, resp.Id)
	return resp
}

type testRequest struct {
	method string
	uri    string
	body   io.Reader
}

func (tr *testRequest) do(handler *HTTPHandler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(tr.method, tr.uri, tr.body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	handler.routes.ServeHTTP(rec, req)
	return rec
}

func TestHandlerHealth(t *testing.T) {
	tr := testRequest{
		method: http.MethodGet,
		uri:    "/health",
	}

	clock := initTestClock()
	cfg := config.Config{}
	queueName := gofakeit.UUID()
	svc := initTestService(t, queueName, clock)
	handler := initTestHandler(cfg, svc)

	rec := tr.do(handler)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandlerSaveJob(t *testing.T) {
	ctx := context.Background()
	objectId := newTestObjectId()
	now := utils.TimeNow()

	t.Run("save new object successfully", func(t *testing.T) {
		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(fmt.Sprintf(`{"object_id": %d}`, objectId)),
		}

		clock := initTestClock()
		cfg := config.Config{}
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)
		handler := initTestHandler(cfg, svc)

		clock.Set(now)
		rec := tr.do(handler)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp Job
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotZero(t, resp.Id)
	})

	t.Run("save same object_id in five minutes, return same job", func(t *testing.T) {
		clock := initTestClock()

		cfg := config.Config{}
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)

		clock.Set(now)
		job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		handler := initTestHandler(cfg, svc)

		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(fmt.Sprintf(`{"object_id": %d}`, job.ObjectId)),
		}

		rec := tr.do(handler)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp Job
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, resp.Id, job.Id)
	})

	t.Run("save same object_id before five minutes, return new job", func(t *testing.T) {
		clock := initTestClock()
		cfg := config.Config{}
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)
		handler := initTestHandler(cfg, svc)

		clock.Set(now.Add(-60 * time.Minute))
		job1, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
		require.NoError(t, err)

		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(fmt.Sprintf(`{"object_id": %d}`, job1.ObjectId)),
		}

		clock.Set(now)

		rec := tr.do(handler)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp Job
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotZero(t, resp.Id)
		assert.NotEqual(t, job1.Id, resp.Id)
	})

	t.Run("invalid object_id", func(t *testing.T) {
		tr := testRequest{
			method: http.MethodPost,
			uri:    fmt.Sprintf("/v1/jobs"),
			body:   strings.NewReader(`{"object_id": abc}`),
		}

		clock := initTestClock()
		cfg := config.Config{}
		queueName := gofakeit.UUID()
		svc := initTestService(t, queueName, clock)
		handler := initTestHandler(cfg, svc)

		rec := tr.do(handler)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandlerGetJob(t *testing.T) {
	ctx := context.Background()

	clock := initTestClock()
	queueName := gofakeit.UUID()
	svc := initTestService(t, queueName, clock)

	job, err := svc.SaveJob(ctx, JobPayload{ObjectId: newTestObjectId()})
	require.NoError(t, err)

	testcases := []struct {
		name       string
		id         string
		statusCode int
	}{
		{
			name:       "get exist job",
			id:         strconv.Itoa(job.Id),
			statusCode: http.StatusOK,
		},
		{
			name:       "get non exist job",
			id:         "99999",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "invalid job id",
			id:         "abc",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tr := testRequest{
				method: http.MethodGet,
				uri:    fmt.Sprintf("/v1/jobs/%s", tc.id),
			}

			clock := initTestClock()
			cfg := config.Config{}
			queueName := gofakeit.UUID()
			svc := initTestService(t, queueName, clock)
			handler := initTestHandler(cfg, svc)

			rec := tr.do(handler)
			assert.Equal(t, tc.statusCode, rec.Code)
		})
	}
}
