package service

import (
	"context"
	answerv1 "github.com/MuxiKeStack/be-api/gen/proto/answer/v1"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	feedv1 "github.com/MuxiKeStack/be-api/gen/proto/feed/v1"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-stance/domain"
	"github.com/MuxiKeStack/be-stance/events"
	"github.com/MuxiKeStack/be-stance/pkg/logger"
	"github.com/MuxiKeStack/be-stance/repository"
	"strconv"
	"time"
)

type StanceService interface {
	Endorse(ctx context.Context, ubs domain.UserBizStance) error
	GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
	CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error)
}

type stanceService struct {
	repo       repository.StanceRepository
	producer   events.Producer
	uidGetters map[stancev1.Biz]UIDGetter
	l          logger.Logger
}

func NewStanceService(repo repository.StanceRepository, producer events.Producer, evaluationClient evaluationv1.EvaluationServiceClient,
	answerClient answerv1.AnswerServiceClient, l logger.Logger) StanceService {
	return &stanceService{
		repo:     repo,
		producer: producer,
		uidGetters: map[stancev1.Biz]UIDGetter{
			stancev1.Biz_Evaluation: &EvaluationUIDGetter{evaluationClient: evaluationClient},
			stancev1.Biz_Answer:     &AnswerUIDGetter{answerClient: answerClient},
		},
		l: l,
	}
}

func (s *stanceService) CountStance(ctx context.Context, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	return s.repo.CountStance(ctx, biz, bizId)
}

func (s *stanceService) Endorse(ctx context.Context, ubs domain.UserBizStance) error {
	err := s.repo.Endorse(ctx, ubs)
	if err != nil {
		return err
	}
	// TODO 有很明显的步骤，可以采用责任链模式，但是目前就两步，真没必要
	// 责任链倾向于同步调用，事件驱动和责任链很像，但是倾向于异步调用，区别在于每一步的执行都是被上一步的消息驱动的，
	// 也就是每一步执行之后都会向下一步的topic里发消息，
	if ubs.Stance == stancev1.Stance_Support {
		// 发送一条支持事件
		go func() {
			// 下面有可能发生一些错误，一旦发生错误这些提醒就被丢失了
			// 如果严格数据一致性，不丢失提醒的话要开分布式事务
			// 相比于引入分布式事务，这里丢一点消息提醒也无伤大雅，核心消息并没有出现问题
			// 而且也可以将失败的任务，引入其他容错机制，但是我这里没有做
			// 去获取被支持者的uid，
			getter, ok := s.uidGetters[ubs.Biz]
			if !ok {
				s.l.Error("biz getter不存在",
					logger.String("biz", ubs.Biz.String()),
					logger.Int64("bizId", ubs.BizId))
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			supported, er := getter.GetUID(ctx, ubs.BizId)
			if er != nil {
				s.l.Error("被支持的biz不合法",
					logger.Error(er),
					logger.String("biz", ubs.Biz.String()),
					logger.Int64("bizId", ubs.BizId))
				return
			}
			er = s.producer.ProduceFeedEvent(ctx, events.FeedEvent{
				Type: feedv1.EventType_Support,
				Metadata: map[string]string{
					"supporter": strconv.FormatInt(ubs.Uid, 10),
					"supported": strconv.FormatInt(supported, 10),
					"biz":       ubs.Biz.String(),
					"bizId":     strconv.FormatInt(ubs.BizId, 10),
				},
			})
			if er != nil {
				s.l.Error("发送支持事件失败",
					logger.Error(er),
					logger.String("biz", ubs.Biz.String()),
					logger.Int64("bizId", ubs.BizId),
					logger.Int64("supporter", ubs.Uid),
					logger.Int64("supported", supported))

			}
		}()

	}
	return nil
}

func (s *stanceService) GetUserStance(ctx context.Context, uid int64, biz stancev1.Biz, bizId int64) (domain.UserBizStance, error) {
	return s.repo.GetUserBizStance(ctx, uid, biz, bizId)
}
