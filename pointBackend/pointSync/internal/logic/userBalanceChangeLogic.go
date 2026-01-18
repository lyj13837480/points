package logic

import (
	"context"
	"pointSync/pointSync/internal/model"
	"pointSync/pointSync/internal/stores/gdb"
)

type UserBalanceChangeLogic struct {
}

var UserBalanceChangeLogicInstance = new(UserBalanceChangeLogic)

func (l *UserBalanceChangeLogic) CreateUserBalanceChange(ctx context.Context, userBalanceChange *model.UserBalanceChange) error {
	return gdb.DB.WithContext(ctx).Model(&model.UserBalanceChange{}).Create(userBalanceChange).Error
}

func (l *UserBalanceChangeLogic) batchUserBalanceChangeLogic(ctx context.Context, userBalanceChanges []model.UserBalanceChange) error {
	return gdb.DB.WithContext(ctx).Model(&model.UserBalanceChange{}).Create(&userBalanceChanges).Error
}
