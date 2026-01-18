package model

import (
	"time"
)

// User 用户表
type User struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	UserAddress string    `gorm:"column:user_address;type:varchar(42);not null;uniqueIndex:uk_user_address;comment:用户地址" json:"user_address"`
	Status      int       `gorm:"column:status;not null;default:1;index:idx_status;comment:用户状态 1:正常 0:禁用" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}
