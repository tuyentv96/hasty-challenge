package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/adjust/rmq/v4"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	migration "github.com/tuyentv96/hasty-challenge/db"
)

func SetupDBTest() (DockerDBConn *pg.DB, closeFunc func() error) {
	cfg := struct {
		Address  string `json:"addr"`
		Database string `json:"db"`
		Username string `json:"user"`
		Password string `json:"password"`
	}{
		Database: "postgres",
		Username: "postgres",
		Password: "123456",
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	runCfg := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12.8-alpine",
		Cmd: []string{
			"postgres",
			"-c", "log_statement=all",
			"-c", "log_destination=stderr",
		},
		Env: []string{
			"POSTGRES_USER=" + cfg.Username,
			"POSTGRES_PASSWORD=" + cfg.Password,
			"POSTGRES_DB=" + cfg.Database,
			"listen_addresses = '*'",
		},
	}

	resource, err := pool.RunWithOptions(runCfg, func(hostConfig *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		hostConfig.AutoRemove = true
		hostConfig.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	closeFunc = resource.Close

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(1000)
	handleInterrupt(pool, resource)

	cfg.Address = resource.Container.NetworkSettings.IPAddress

	// Docker layer network is different on Mac
	cfg.Address = net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))

	dataSource := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Address,
		cfg.Database,
	)

	if err := pool.Retry(func() error {
		DockerDBConn = pg.Connect(&pg.Options{
			Addr:     cfg.Address,
			Database: cfg.Database,
			User:     cfg.Username,
			Password: cfg.Password,
		})

		_, err := DockerDBConn.ExecContext(context.Background(), "SELECT 1")
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := migration.Migrate(dataSource); err != nil {
		log.Println(err.Error())
	}

	return
}

func handleInterrupt(pool *dockertest.Pool, container *dockertest.Resource) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := pool.Purge(container); err != nil {
			log.Fatalf("Could not purge container: %s", err)
		}
		os.Exit(0)
	}()
}

func NewConnection(db *redis.Client, tag string) (rmq.Connection, error) {
	return rmq.OpenConnectionWithRedisClient(tag, db, nil)
}

func NewQueue(conn rmq.Connection, queueName string) (rmq.Queue, error) {
	return conn.OpenQueue(queueName)
}

func SetupRedisTest() (redisClient *redis.Client, closeFunc func() error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	runCfg := &dockertest.RunOptions{
		Repository: "redis",
		Tag:        "5.0.13-alpine3.14",
	}

	resource, err := pool.RunWithOptions(runCfg, func(hostConfig *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		hostConfig.AutoRemove = true
		hostConfig.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	closeFunc = resource.Close
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(1000)
	handleInterrupt(pool, resource)

	address := net.JoinHostPort("localhost", resource.GetPort("6379/tcp"))

	if err := pool.Retry(func() error {
		redisClient = redis.NewClient(&redis.Options{
			Addr: address,
		})

		return redisClient.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return
}
