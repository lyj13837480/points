package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// CalculatedPoints 计算出的积分JSON
type CalculatedPoints map[string]interface{}

// Value 实现driver.Valuer接口
func (c CalculatedPoints) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现sql.Scanner接口
func (c *CalculatedPoints) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// UserPointsChange 用户积分变动记录表
type UserPointsChange struct {
	ID               int64            `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID           int              `gorm:"column:user_id;not null;index:idx_user_id;comment:用户ID" json:"user_id"`
	ChainID          int              `gorm:"column:chain_id;not null;index:idx_chain_id;comment:链ID" json:"chain_id"`
	FromBlockHeight  int64            `gorm:"column:from_block_height;not null;comment:变动前区块高度" json:"from_block_height"`
	ToBlockHeight    int64            `gorm:"column:to_block_height;not null;comment:变动后区块高度" json:"to_block_height"`
	FromBlockTime    int64            `gorm:"column:from_block_time;not null;comment:变动前区块时间" json:"from_block_time"`
	ToBlockTime      int64            `gorm:"column:to_block_time;not null;comment:变动后区块时间" json:"to_block_time"`
	CalculatedPoints CalculatedPoints `gorm:"column:calculated_points;type:json;not null;comment:计算出的积分JSON" json:"calculated_points"`
	Status           int              `gorm:"column:status;not null;default:1;index:idx_status;comment:状态 1:成功 0:失败" json:"status"`
	Reason           string           `gorm:"column:reason;type:varchar(255);comment:失败原因" json:"reason"`
	CreatedAt        time.Time        `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (UserPointsChange) TableName() string {
	return "user_points_change"
}
