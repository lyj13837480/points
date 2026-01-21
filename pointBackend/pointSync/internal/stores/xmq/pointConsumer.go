package xmq

import (
	"context"
	"encoding/json"
	"pointSync/internal/model"
	"pointSync/internal/service/point"

	"github.com/zeromicro/go-zero/core/logx"
)

type PointConsumer struct {
	ctx context.Context
}

func NewPointConsumer(ctx context.Context) *PointConsumer {
	return &PointConsumer{
		ctx: ctx,
	}
}

func (l *PointConsumer) Consume(ctx context.Context, key, value string) error {
	// 解析消息
	var msg []model.UserBalanceID
	err := json.Unmarshal([]byte(value), &msg)
	if err != nil {
		logx.Errorf("Failed to unmarshal point message: %v", err)
		return err
	}
	for i := 0; i < len(msg); i++ {
		point.CalPoint(l.ctx, msg[i].ID)
	}

	logx.Infof("PointConsumer key :%s , value :%s", key, value)
	return nil
}
