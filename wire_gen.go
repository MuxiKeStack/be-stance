// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/MuxiKeStack/be-stance/grpc"
	"github.com/MuxiKeStack/be-stance/ioc"
	"github.com/MuxiKeStack/be-stance/pkg/grpcx"
	"github.com/MuxiKeStack/be-stance/repository"
	"github.com/MuxiKeStack/be-stance/repository/cache"
	"github.com/MuxiKeStack/be-stance/repository/dao"
	"github.com/MuxiKeStack/be-stance/service"
)

// Injectors from wire.go:

func InitGRPCServer() grpcx.Server {
	logger := ioc.InitLogger()
	db := ioc.InitDB(logger)
	stanceDAO := dao.NewGORMStanceDAO(db)
	cmdable := ioc.InitRedis()
	stanceCache := cache.NewRedisStanceCache(cmdable)
	stanceRepository := repository.NewCachedStanceRepository(stanceDAO, stanceCache, logger)
	stanceService := service.NewStanceService(stanceRepository)
	stanceServiceServer := grpc.NewStanceServiceServer(stanceService)
	client := ioc.InitEtcdClient()
	server := ioc.InitGRPCxKratosServer(stanceServiceServer, client, logger)
	return server
}