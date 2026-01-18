package logic

import (
	"context"
	"pointSync/pointSync/internal/model"
	"pointSync/pointSync/internal/stores/gdb"
)

type UserLogic struct {
}

var UserLogicInstance = new(UserLogic)

func (l *UserLogic) NewUserLogic(ctx context.Context, userModel *model.User) error {
	return gdb.DB.WithContext(ctx).Model(&model.User{}).Create(&userModel).Error
}

func (l *UserLogic) GetUserByAddress(ctx context.Context, userAddress string) (userModel model.User, err error) {
	err = gdb.DB.WithContext(ctx).Model(&model.User{}).Where("user_address = ? and  status <> 0", userAddress).First(&userModel).Error
	if userModel.ID == 0 {
		userModel.UserAddress = userAddress
		userModel.Status = 1
		err = l.NewUserLogic(ctx, &userModel)
	}
	return
}
