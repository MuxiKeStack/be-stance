package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrDuplicateOperation = errors.New("operation is duplicate")
	ErrRecordNotFound     = gorm.ErrRecordNotFound
)

type StanceDAO interface {
	Upsert(ctx context.Context, userBizStance UserBizStance) (Increment, error)
	GetUserBizStance(ctx context.Context, uid int64, biz int32, bizId int64) (UserBizStance, error)
	GetBizStanceCount(ctx context.Context, biz int32, bizId int64) (BizStanceCount, error)
}

type GORMStanceDAO struct {
	db *gorm.DB
}

func NewGORMStanceDAO(db *gorm.DB) StanceDAO {
	return &GORMStanceDAO{db: db}
}

type Increment struct {
	Support int64
	Oppose  int64
}

func (dao *GORMStanceDAO) Upsert(ctx context.Context, userBizStance UserBizStance) (Increment, error) {
	now := time.Now().UnixMilli()
	var increment Increment
	return increment, dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 因为计数的判断要根据前后的结果来判断
		var ubs UserBizStance
		err := tx.Where("uid = ? and biz = ? and biz_id = ?",
			userBizStance.Uid, userBizStance.Biz, userBizStance.BizId).
			First(&ubs).Error
		switch {
		case err == nil:
			// 更新
			if ubs.Stance == userBizStance.Stance {
				return ErrDuplicateOperation
			}
			er := tx.Model(&UserBizStance{}).
				Where("uid = ? and biz = ? and biz_id = ?",
					userBizStance.Uid, userBizStance.Biz, userBizStance.BizId).
				Updates(map[string]any{
					"utime":  now,
					"stance": userBizStance.Stance,
				}).Error
			if er != nil {
				return er
			}
			increment = dao.GetIncrement(ubs.Stance, userBizStance.Stance)
			// 更新计数
			return tx.Model(&BizStanceCount{}).
				Where("biz = ? and biz_id = ?",
					userBizStance.Biz, userBizStance.BizId).
				Updates(map[string]any{
					"utime": now,
					// 这里是可以 + -1吗
					"support_cnt": gorm.Expr("`support_cnt` + ?", increment.Support),
					"oppose_cnt":  gorm.Expr("`oppose_cnt` + ?", increment.Oppose),
				}).Error
		case err == gorm.ErrRecordNotFound:
			userBizStance.Ctime = now
			userBizStance.Utime = now
			er := tx.Create(&userBizStance).Error
			if er != nil {
				return er
			}
			increment = dao.GetIncrement(0, userBizStance.Stance)
			// 更新计数
			return tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"utime":       now,
					"support_cnt": gorm.Expr("`support_cnt` + ?", increment.Support),
					"oppose_cnt":  gorm.Expr("`oppose_cnt` + ?", increment.Oppose),
				}),
			}).Create(&BizStanceCount{
				Biz:        userBizStance.Biz,
				BizId:      userBizStance.BizId,
				SupportCnt: increment.Support,
				OpposeCnt:  increment.Oppose,
			}).Error
		default:
			return err
		}
	})
}

func (dao *GORMStanceDAO) GetIncrement(oldStance, newStance int32) Increment {
	var increment Increment

	if oldStance == 1 {
		increment.Support = -1
	} else if oldStance == -1 {
		increment.Oppose = -1
	}

	if newStance == 1 {
		increment.Support += 1
	} else if newStance == -1 {
		increment.Oppose += 1
	}

	return increment
}

func (dao *GORMStanceDAO) GetUserBizStance(ctx context.Context, uid int64, biz int32, bizId int64) (UserBizStance, error) {
	var ubs UserBizStance
	err := dao.db.WithContext(ctx).
		Where("uid = ? and biz = ? and biz_id = ?", uid, biz, bizId).
		First(&ubs).Error
	return ubs, err
}

func (dao *GORMStanceDAO) GetBizStanceCount(ctx context.Context, biz int32, bizId int64) (BizStanceCount, error) {
	var bsc BizStanceCount
	err := dao.db.WithContext(ctx).
		Where("biz = ? and biz_id = ?", biz, bizId).
		First(&bsc).Error
	return bsc, err
}

type UserBizStance struct {
	Id     int64 `gorm:"primaryKey,autoIncrement"`
	Uid    int64 `gorm:"uniqueIndex:id_biz_bizId"`
	Biz    int32 `gorm:"uniqueIndex:id_biz_bizId"`
	BizId  int64 `gorm:"uniqueIndex:id_biz_bizId"`
	Stance int32
	Utime  int64
	Ctime  int64
}

type BizStanceCount struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Biz        int32 `gorm:"uniqueIndex:id_biz_bizId"`
	BizId      int64 `gorm:"uniqueIndex:id_biz_bizId"`
	SupportCnt int64
	OpposeCnt  int64
	Utime      int64
	Ctime      int64
}
