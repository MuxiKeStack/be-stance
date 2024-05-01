//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-stance/grpc"
	"github.com/MuxiKeStack/be-stance/ioc"
	"github.com/MuxiKeStack/be-stance/pkg/grpcx"
	"github.com/MuxiKeStack/be-stance/repository"
	"github.com/MuxiKeStack/be-stance/repository/cache"
	"github.com/MuxiKeStack/be-stance/repository/dao"
	"github.com/MuxiKeStack/be-stance/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewStanceServiceServer,
		service.NewStanceService,
		// producer
		ioc.InitProducer,
		// rpc client
		ioc.InitAnswerClient, ioc.InitEvaluationClient,
		repository.NewCachedStanceRepository,
		dao.NewGORMStanceDAO, cache.NewRedisStanceCache,
		// 第三方
		ioc.InitKafka,
		ioc.InitDB, ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitLogger,
	)
	return grpcx.Server(nil)
}
