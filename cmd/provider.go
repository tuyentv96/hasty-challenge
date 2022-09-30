package cmd

import (
	"github.com/adjust/rmq/v5"
	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/jobs"
	"github.com/tuyentv96/hasty-challenge/utils"
)

func ProvideConfig() (config.Config, error) {
	cfg := config.Config{}
	godotenv.Load()

	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func ProvideLogger(cfg config.Config) *logrus.Entry {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LoggerConfig.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logger.Level = level
	return logrus.NewEntry(logger)
}

func ProvideClock() clock.Clock {
	return clock.New()
}

func ProvideRandom() utils.Random {
	return utils.NewRandomImpl()
}

func ProvidePostgres(cfg config.Config) (*pg.DB, error) {
	db := pg.Connect(&pg.Options{
		Addr:     cfg.SQLConfig.SQLAddress,
		User:     cfg.SQLConfig.SQLUser,
		Password: cfg.SQLConfig.SQLPassword,
		Database: cfg.SQLConfig.SQLName,
	})

	if _, err := db.ExecOne("SELECT 1"); err != nil {
		return nil, err
	}

	return db, nil
}

func ProvideTransactioner(db *pg.DB) utils.Transactioner {
	return utils.NewTransaction(db)
}

func ProvideJobSvc(jobStore jobs.Store, queue rmq.Queue, clock clock.Clock) jobs.Service {
	return jobs.NewService(jobStore, queue, clock)
}

func ProvideJobStore(db *pg.DB) jobs.Store {
	return jobs.NewJobStore(db)
}

func ProvideJobHandler(cfg config.Config, jobSvc jobs.Service) *jobs.HTTPHandler {
	return jobs.NewHTTPHandler(cfg, jobSvc)
}

func ProvideJobWorker(cfg config.Config, logger *logrus.Entry, jobSvc jobs.Service, connection rmq.Connection, queue rmq.Queue, clock clock.Clock, random utils.Random, transactioner utils.Transactioner) jobs.Worker {
	return jobs.NewWorker(cfg, logger, jobSvc, connection, queue, clock, random, transactioner)
}

func ProvideRedis(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.RedisAddress,
		Password: cfg.RedisConfig.RedisPassword,
	})
}

func ProvideRmqConnection(redisClient *redis.Client) (rmq.Connection, func(), error) {
	closeChan := make(chan bool)
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan, closeChan)

	conn, err := rmq.OpenConnectionWithRedisClient(jobs.QueueName, redisClient, errChan)
	if err != nil {
		return nil, func() {}, err
	}

	return conn, func() {
		close(closeChan)
		conn.StopAllConsuming()
	}, nil
}

func ProvideRedisQueue(conn rmq.Connection) (rmq.Queue, func(), error) {
	queue, err := conn.OpenQueue(jobs.QueueName)
	if err != nil {
		return nil, func() {}, err
	}

	return queue, func() {
		queue.StopConsuming()
	}, nil
}

func rmqLogErrors(errChan <-chan error, closeChan <-chan bool) {
	for {
		select {
		case <-closeChan:
			return
		case err := <-errChan:
			switch err := err.(type) {
			case *rmq.HeartbeatError:
				if err.Count == rmq.HeartbeatErrorLimit {
					logrus.WithError(err).Error("[rmq] heartbeat error (limit)")
				} else {
					logrus.WithError(err).Error("[rmq] heartbeat error")
				}
			case *rmq.ConsumeError:
				logrus.WithError(err).Error("[rmq] consume error")
			case *rmq.DeliveryError:
				logrus.WithError(err).Error("[rmq] delivery error")
			default:
				logrus.WithError(err).Error("[rmq] other error")
			}
		}
	}
}
