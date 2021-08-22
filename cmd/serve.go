package cmd

import "github.com/urfave/cli"

// Serve creates a command that start an http server
func (a *ApplicationContext) Serve() cli.Command {
	return cli.Command{
		Name:  "serve",
		Usage: "serve http request",
		Action: func(c *cli.Context) error {
			a.jobHandler.Serve()
			return nil
		},
	}
}
