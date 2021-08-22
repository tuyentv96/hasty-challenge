//+build wireinject

package cmd

import (
	"context"

	"github.com/google/wire"
)

var ApplicationSet = wire.NewSet(
	ProvideConfig,
	ProvideLogger,
	ProvideClock,
	ProvideRandom,
	ProvidePostgres,
	ProvideTransactioner,
	ProvideRedis,
	ProvideRmqConnection,
	ProvideRedisQueue,

	ProvideJobSvc,
	ProvideJobStore,
	ProvideJobHandler,
	ProvideJobWorker,
)

func InitApplication(ctx context.Context) (*ApplicationContext, func(), error) {
	wire.Build(
		ApplicationSet,
		wire.Struct(new(ApplicationContext), "*"),
	)

	return nil, nil, nil
}
