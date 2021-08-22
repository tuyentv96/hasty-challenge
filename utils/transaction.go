package utils

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type transactionKey struct{}

type Transactioner interface {
	RunWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Transaction struct {
	db *pg.DB
}

func NewTransaction(db *pg.DB) *Transaction {
	return &Transaction{
		db: db,
	}
}

func (t *Transaction) RunWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.WithContext(ctx).Begin()
	if err != nil {
		return err
	}

	// https://pseudomuto.com/2018/01/clean-sql-transactions-in-golang/
	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	ctx = context.WithValue(ctx, transactionKey{}, tx)
	err = fn(ctx)

	return err
}

func TransactionFromContext(ctx context.Context, fallback orm.DB) orm.DB {
	val := ctx.Value(transactionKey{})
	if tx, ok := val.(orm.DB); ok {
		return tx
	}

	return fallback
}
