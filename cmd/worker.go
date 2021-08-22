package cmd

import "github.com/urfave/cli"

func (a *ApplicationContext) Worker() cli.Command {
	return cli.Command{
		Name:  "worker",
		Usage: "worker",
		Action: func(c *cli.Context) error {
			go a.jobWorker.RunCleaner()
			return a.jobWorker.Start()
		},
	}
}
