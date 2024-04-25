package service

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-stance/domain"
	"github.com/MuxiKeStack/be-stance/repository"
)

type StanceService interface {
	Endorse(ctx context.Context, ubs domain.UserBizStance) error
	GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
	CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
}

type stanceService struct {
	repo repository.StanceRepository
}

func (s *stanceService) CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	return s.repo.CountStance(ctx, biz, bizId)
}

func NewStanceService(repo repository.StanceRepository) StanceService {
	return &stanceService{repo: repo}
}

func (s *stanceService) Endorse(ctx context.Context, ubs domain.UserBizStance) error {
	return s.repo.Endorse(ctx, ubs)
}

func (s *stanceService) GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	return s.repo.GetUserBizStance(ctx, uid, biz, bizId)
}
