package logic

import (
	"context"
	"pointSync/internal/model"
	"pointSync/internal/stores/gdb"
)

type UserLogic struct {
}

var UserLogicInstance = new(UserLogic)

func (l *UserLogic) NewUserLogic(ctx context.Context, userModel *model.User) error {
	return gdb.DB.WithContext(ctx).Model(&model.User{}).Create(&userModel).Error
}

func (l *UserLogic) GetUserByAddress(ctx context.Context, userAddress string) (userModel model.User, err error) {
	err = gdb.DB.WithContext(ctx).Model(&model.User{}).Where("user_address = ? and  status <> 0", userAddress).First(&userModel).Error
	return
}
