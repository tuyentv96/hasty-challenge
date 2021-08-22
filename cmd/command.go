package cmd

import (
	"context"

	"github.com/urfave/cli"

	"github.com/tuyentv96/hasty-challenge/config"
	"github.com/tuyentv96/hasty-challenge/jobs"
)

type ApplicationContext struct {
	ctx        context.Context
	cfg        config.Config
	jobStore   jobs.Store
	jobSvc     jobs.Service
	jobHandler *jobs.HTTPHandler
	jobWorker  jobs.Worker
}

func (a *ApplicationContext) Commands() *cli.App {
	app := cli.NewApp()
	app.Before = func(c *cli.Context) error {
		if err := a.Migrate(); err != nil {
			return err
		}

		return nil
	}

	app.Commands = []cli.Command{
		a.Serve(),
		a.Worker(),
	}

	return app
}
