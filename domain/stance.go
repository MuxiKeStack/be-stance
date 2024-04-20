package domain

import stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"

type UserBizStance struct {
	Uid        int64
	Biz        stancev1.Biz
	BizId      int64
	Stance     stancev1.Stance
	SupportCnt int64
	OpposeCnt  int64
}
