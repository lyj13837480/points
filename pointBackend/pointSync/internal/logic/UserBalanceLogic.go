package logic

import (
	"context"
	"math/big"
	"pointSync/pointSync/internal/model"
	"pointSync/pointSync/internal/stores/gdb"
)

type UserBalanceLogic struct {
}

var UserBalanceLogicInstance = new(UserBalanceLogic)

func (l *UserBalanceLogic) Save(ctx context.Context, userID int64, chainID int64, balance *big.Int, lastUpdatedBlock int64) error {
	ub, err := l.GetUserBalance(ctx, userID, chainID)
	if err != nil {
		return err
	}
	ub.Balance.Add(ub.Balance, balance)
	ub.LastUpdatedBlock = lastUpdatedBlock
	ub.UserID = userID
	ub.ChainID = chainID
	return gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", userID, chainID).Save(&ub).Error
}

func (l *UserBalanceLogic) GetUserBalance(ctx context.Context, userID int64, chainID int64) (userBalance model.UserBalance, err error) {
	err = gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", userID, chainID).First(&userBalance).Error
	return
}
