package point

import (
	"context"
	"encoding/json"
	"log"
	"pointSync/internal/config"
	"pointSync/internal/logger/xzap"
	"pointSync/internal/model"
	"pointSync/internal/stores/gdb"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

const (
	PointTopic     = "point-topic"
	SecondPerPoint = 0.05            // 每 token 每分钟可产生的积分（例如 0.05）
	SleepTime      = 5 * time.Second // 每次计算间隔时间（）
)

type Service struct {
	KqPusherClient *kq.Pusher
	Ctx            context.Context
	PushConf       *config.KqConf
	Db             *gorm.DB
}

// BalanceSnap 余额快照
type BalanceSnap struct {
	Ts     time.Time       // 快照时间
	Amount decimal.Decimal // 当时余额（token 数）
}

func New(kqPusherClient *kq.Pusher, ctx context.Context, pushConf *config.KqConf, db *gorm.DB) *Service {
	return &Service{
		KqPusherClient: kqPusherClient,
		Ctx:            ctx,
		PushConf:       pushConf,
		Db:             db,
	}
}

func (s *Service) Start() {
	threading.GoSafe(s.PushPointTask)
}

// PushPointTask 推送积分任务
func (s *Service) PushPointTask() {
	for {
		select {
		case <-s.Ctx.Done():
			xzap.WithContext(s.Ctx).Info("PushPointTask ctx done")
			return
		default:
		}
		currentTime := time.Now().Unix() //先简化取服务器当前时间，可调整为取链上最新时间
		// 从数据库中查询需要计算的用户
		var count, i int64
		s.Db.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("status = 1 and (last_update_point_time < ? OR last_update_point_time IS NULL)", currentTime).Count(&count)

		// 分页查询用户并推送 可考虑拆解为多线程执行 当前为单线程循环执行
		for i = 0; i < count; i += s.PushConf.UserLimit {
			var userBalanceIds []model.UserBalanceID
			s.Db.WithContext(s.Ctx).Model(&model.UserBalance{}).Select("id").Where("status = 1 and (last_update_point_time < ? OR last_update_point_time IS NULL)", currentTime).
				Limit(int(s.PushConf.UserLimit)).Offset(int(i)).Find(&userBalanceIds)
			// 支持批用户推送
			for j := 0; j < len(userBalanceIds); j += int(s.PushConf.UserMaxCount) {
				m := j + int(s.PushConf.UserMaxCount)
				if m > len(userBalanceIds) {
					m = len(userBalanceIds)
				}
				ids := userBalanceIds[j:m]
				body, err := json.Marshal(ids)
				if err != nil {
					log.Fatal(err)
				}
				s.KqPusherClient.Push(s.Ctx, string(body))
			}
		}
		time.Sleep(SleepTime)
	}
}

/**
 * CalPoint 计算用户积分
 * @param ctx context.Context
 * @param id uint64 用户ID
 * @param db *gorm.DB 数据库连接
 * @return error
 */
func CalPoint(ctx context.Context, id uint64) error {
	var userBalance model.UserBalance
	err := gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("id = ?", id).First(&userBalance).Error
	if err != nil {
		return err
	}
	// 计算积分
	// 1. 从 last_update_point_time 到 currentTime 这段时间内，用户的余额变化了多少（假设为 delta）
	// 2. 积分 = delta * 系数 * 分钟数
	// 3. 更新 user_balance 表的 points 字段（累加积分）和 last_update_point_time 字段（设为 currentTime）
	// 1. 从 last_update_point_time 到 currentTime 这段时间内，用户的余额变化了多少（假设为 delta）
	currentTime := time.Now()
	lastUpdatePointTime := time.Unix(userBalance.LastUpdatePointTime, 0)

	// 2. 从 user_balance_change 表中查询用户在这段时间内的余额变化记录（假设为 N 条）
	// 3. 按时间升序排序，得到 N+1 条快照（起始快照 + N 条变化快照）
	var userBalanceChange []model.UserBalanceChange
	err = gdb.DB.WithContext(ctx).Model(&model.UserBalanceChange{}).Where("user_id = ? and chain_id = ? and block_time >= ? and block_time < ?", userBalance.UserID, userBalance.ChainID, lastUpdatePointTime.Unix(), currentTime.Unix()).Order("block_time asc").Find(&userBalanceChange).Error
	if err != nil {
		return err
	}
	var snaps []BalanceSnap
	//如果余额变化表没有记录，说明用户在这段时间内没有余额变化，直接用最后一次更新时间的余额作为起始快照
	if len(userBalanceChange) == 0 {
		snaps = append(snaps, BalanceSnap{Ts: time.Unix(userBalance.LastUpdatedBlock, 0), Amount: userBalance.Balance})
	}
	for _, change := range userBalanceChange {
		snaps = append(snaps, BalanceSnap{Ts: time.Unix(change.BlockTime, 0), Amount: change.BalanceAfter})
	}
	// 4. 调用 CalcPoints 函数计算积分
	points := CalcPoints(SecondPerPoint, lastUpdatePointTime, currentTime, snaps)
	// 5. 更新 user_balance 表的 points 字段（累加积分）和 last_update_point_time 字段（设为 currentTime）
	userBalance.Points = userBalance.Points.Add(points)
	userBalance.LastUpdatePointTime = currentTime.Unix()
	err = gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"points":                 userBalance.Points,
		"last_update_point_time": userBalance.LastUpdatePointTime,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// CalcPoints 计算积分
// coeff：每 token 每分钟可产生的积分（例如 0.05）
// from, to：要统计的左闭右开区间
// snaps：必须按时间升序排列，且 from 之前最好有一条“起始快照”（余额为 0 也可以）
func CalcPoints(coeff float64, from, to time.Time, snaps []BalanceSnap) decimal.Decimal {
	if len(snaps) == 0 || from.Equal(to) || to.Before(from) {
		return decimal.NewFromFloat(0)
	}

	// 1. 裁剪掉区间外的快照
	var valid []BalanceSnap
	for _, s := range snaps {
		if s.Ts.Before(from) {
			continue
		}
		if s.Ts.After(to) {
			break
		}
		valid = append(valid, s)
	}

	// 2. 如果区间起点没有快照，用最近一条“from 之前”的快照补齐
	if len(valid) == 0 || valid[0].Ts.After(from) {
		// 找 from 之前最近一条
		var last BalanceSnap
		for i := len(snaps) - 1; i >= 0; i-- {
			if !snaps[i].Ts.After(from) {
				last = snaps[i]
				break
			}
		}
		last.Ts = from
		valid = append([]BalanceSnap{last}, valid...)
	}

	// 3. 如果区间终点没有快照，用最后一条补齐
	if lastValid := valid[len(valid)-1]; lastValid.Ts.Before(to) {
		valid = append(valid, BalanceSnap{Ts: to, Amount: lastValid.Amount})
	}

	// 4. 分段累加
	var points decimal.Decimal
	for i := 1; i < len(valid); i++ {
		prev := valid[i-1]
		curr := valid[i]

		durMin := curr.Ts.Sub(prev.Ts).Minutes()
		if durMin <= 0 {
			continue
		}
		// 积分 = 余额 * 系数 * 分钟数
		points = points.Add(prev.Amount.Mul(decimal.NewFromFloat(coeff)).Mul(decimal.NewFromFloat(durMin)))
	}
	// 保留 4 位小数，可按业务调整
	points = points.Round(4)
	return points
}
