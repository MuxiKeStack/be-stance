package service

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-stance/domain"
)

type StanceService interface {
	Endorse(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64, stance stancev1.Stance) error
	GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
}

type stanceService struct {
}

func (s *stanceService) Endorse(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64, stance stancev1.Stance) error {
	//TODO implement me
	panic("implement me")
}

func (s *stanceService) GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	//TODO implement me
	panic("implement me")
}
