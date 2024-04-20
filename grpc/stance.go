package grpc

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-stance/service"
)

type StanceServiceServer struct {
	stancev1.UnimplementedStanceServiceServer
	svc service.StanceService
}

func (s *StanceServiceServer) Endorse(ctx context.Context, request *stancev1.EndorseRequest) (*stancev1.EndorseResponse, error) {
	err := s.svc.Endorse(ctx, request.GetUid(), request.GetBiz(), request.GetBizId(), request.GetStance())
	return &stancev1.EndorseResponse{}, err
}

func (s *StanceServiceServer) GetUserStance(ctx context.Context, request *stancev1.GetUserStanceRequest) (*stancev1.GetUserStanceResponse, error) {
	stance, err := s.svc.GetUserStance(ctx, request.GetUid(), request.GetBiz(), request.GetBizId())
	return &stancev1.GetUserStanceResponse{
		Stance:        stance.Stance,
		TotalSupports: stance.SupportCnt,
		TotalOpposes:  stance.OpposeCnt,
	}, err
}
