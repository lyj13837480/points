package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// UserBalance 用户余额表
type UserBalance struct {
	ID                  int64           `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID              int64           `gorm:"column:user_id;not null;index:idx_user_id;comment:用户ID" json:"user_id"`
	ChainID             int64           `gorm:"column:chain_id;not null;index:idx_chain_id;comment:链ID" json:"chain_id"`
	Balance             decimal.Decimal `gorm:"column:balance;not null;comment:余额" json:"balance"`
	LastUpdatedBlock    int64           `gorm:"column:last_updated_block;not null;default:0;comment:已更新区块高度" json:"last_updated_block"`
	CreatedAt           time.Time       `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"updated_at"`
	Points              decimal.Decimal `gorm:"column:points;not null;default:0;comment:积分" json:"points"`
	LastUpdatePointTime int64           `gorm:"column:last_update_point_time;not null;default:0;comment:最后更新积分时间" json:"last_update_point_time"`
}

type UserBalanceID struct {
	ID uint64 `gorm:"column:id"`
}

// TableName 指定表名
func (UserBalance) TableName() string {
	return "user_balance"
}
