package logic

import (
	"context"
	"math/big"
	"pointSync/internal/model"
	"pointSync/internal/stores/gdb"

	"github.com/shopspring/decimal"
)

type UserBalanceLogic struct {
}

var UserBalanceLogicInstance = new(UserBalanceLogic)

func (l *UserBalanceLogic) Save(ctx context.Context, userID int64, chainID int64, balance *big.Int, lastUpdatedBlock int64) error {
	ub, err := l.GetUserBalance(ctx, userID, chainID)
	if err != nil {
		return err
	}
	ub.Balance.Add(decimal.NewFromBigInt(balance, 0))
	ub.LastUpdatedBlock = lastUpdatedBlock
	ub.UserID = userID
	ub.ChainID = chainID
	return gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", userID, chainID).Save(&ub).Error
}

func (l *UserBalanceLogic) GetUserBalance(ctx context.Context, userID int64, chainID int64) (userBalance model.UserBalance, err error) {
	err = gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", userID, chainID).First(&userBalance).Error
	return
}

func (l *UserBalanceLogic) CreateUserBalance(ctx context.Context, userID int64, chainID int64) error {
	ub := &model.UserBalance{
		UserID:           userID,
		ChainID:          chainID,
		Balance:          decimal.NewFromBigInt(big.NewInt(0), 0),
		LastUpdatedBlock: 0,
	}
	return gdb.DB.WithContext(ctx).Model(&model.UserBalance{}).Create(&ub).Error
}
