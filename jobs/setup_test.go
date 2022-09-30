package jobs

import (
	"log"
	"os"
	"testing"

	"github.com/adjust/rmq/v5"
	"github.com/go-pg/pg/v9"
	"github.com/sirupsen/logrus"

	"github.com/tuyentv96/hasty-challenge/utils"
)

var (
	testDb            *pg.DB
	testRmqConnection rmq.Connection
	testStore         Store
	testLogger        *logrus.Entry
	testTransaction   utils.Transactioner
)

func TestMain(m *testing.M) {
	logger := logrus.New()
	testLogger = logrus.NewEntry(logger)

	var redisCloseFunc func() error
	redisClient, redisCloseFunc := utils.SetupRedisTest()

	var err error
	testRmqConnection, err = rmq.OpenConnectionWithRedisClient("test", redisClient, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var closeFunc func() error
	testDb, closeFunc = utils.SetupDBTest()
	testTransaction = utils.NewTransaction(testDb)
	testStore = NewJobStore(testDb)

	code := m.Run()
	closeFunc()
	redisCloseFunc()
	os.Exit(code)
}
