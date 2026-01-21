package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// UserBalanceChange 用户余额变动记录表
type UserBalanceChange struct {
	ID           int64           `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserID       int64           `gorm:"column:user_id;not null;index:idx_user_id;comment:用户ID" json:"user_id"`
	ChainID      int64           `gorm:"column:chain_id;not null;index:idx_chain_id;comment:链ID" json:"chain_id"`
	ChangeType   string          `gorm:"column:change_type;type:varchar(20);not null;comment:变动类型mint、burn、transfer_in、transfer_out" json:"change_type"`
	Amount       decimal.Decimal `gorm:"column:amount;type:decimal(65,0);not null;comment:变动金额正数表示增加，负数表示减少" json:"amount"`
	BalanceAfter decimal.Decimal `gorm:"column:balance_after;type:decimal(65,0);not null;comment:变动后的余额" json:"balance_after"`
	TxHash       string          `gorm:"column:tx_hash;type:varchar(66);not null;index:idx_tx_hash;comment:交易哈希" json:"tx_hash"`
	BlockTime    int64           `gorm:"column:block_time;not null;comment:区块时间" json:"block_time"`
	BlockHeight  int64           `gorm:"column:block_height;not null;index:idx_block_height;comment:区块高度" json:"block_height"`
	CreatedAt    time.Time       `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (UserBalanceChange) TableName() string {
	return "user_balance_change"
}
