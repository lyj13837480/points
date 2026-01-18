package model

import (
	"math/big"
	"time"
)

// UserPoints 用户积分表
type UserPoints struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID           int       `gorm:"column:user_id;not null;index:idx_user_id;comment:用户ID" json:"user_id"`
	ChainID          int       `gorm:"column:chain_id;not null;index:idx_chain_id;comment:链ID" json:"chain_id"`
	Points           *big.Int  `gorm:"column:points;not null;comment:积分" json:"points"`
	LastUpdatedBlock int64     `gorm:"column:last_updated_block;not null;default:0;comment:已更新区块高度" json:"last_updated_block"`
	CreatedAt        time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (UserPoints) TableName() string {
	return "user_points"
}
