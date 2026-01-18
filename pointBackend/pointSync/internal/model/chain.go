package model

import (
	"time"
)

// Chain 区块链配置表
type Chain struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement;comment:主键" json:"id"`
	ChainID             int64     `gorm:"column:chain_id;not null;comment:链ID" json:"chain_id"`
	Name                string    `gorm:"column:name;type:varchar(255);not null;comment:合约名称" json:"name"`
	Symbol              string    `gorm:"column:symbol;type:varchar(10);not null;comment:合约符号" json:"symbol"`
	RPCURL              string    `gorm:"column:rpc_url;type:varchar(255);not null;comment:RPC地址" json:"rpc_url"`
	ContractAddress     string    `gorm:"column:contract_address;type:varchar(42);not null;comment:合约地址" json:"contract_address"`
	IsActive            bool      `gorm:"column:is_active;not null;default:true;comment:是否活跃" json:"is_active"`
	StartBlock          uint64    `gorm:"column:start_block;not null;default:0;comment:开始区块高度" json:"start_block"`
	LastConfirmedBlock  uint64    `gorm:"column:last_confirmed_block;not null;default:0;comment:最后确认的区块高度" json:"last_confirmed_block"`
	LastProcessedBlock  uint64    `gorm:"column:last_processed_block;not null;default:0;comment:最后处理的区块高度" json:"last_processed_block"`
	LastCalculatedBlock uint64    `gorm:"column:last_calculated_block;not null;default:0;comment:最后计算积分的区块高度" json:"last_calculated_block"`
	CreatedAt           time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (Chain) TableName() string {
	return "chain"
}
