package cmd

import (
	"fmt"

	_ "github.com/lib/pq"
	migration "github.com/tuyentv96/hasty-challenge/db"
)

func (a *ApplicationContext) Migrate() error {
	dataSource := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		a.cfg.SQLUser,
		a.cfg.SQLPassword,
		a.cfg.SQLAddress,
		a.cfg.SQLName,
	)

	if err := migration.Migrate(dataSource); err != nil {
		return err
	}

	return nil
}
