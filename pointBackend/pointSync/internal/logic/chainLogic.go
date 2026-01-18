package logic

import (
	"context"
	"pointSync/pointSync/internal/model"
	"pointSync/pointSync/internal/stores/gdb"
)

type ChainLogic struct {
	ctx context.Context
}

var ChainLogicInstance = new(ChainLogic)

/***
 * @description: 查询链数据
 * @param {context.Context} ctx
 * @param {int} chainID
 * @return {*}
 */

func (l *ChainLogic) QueryChainData(ctx context.Context) (chains []model.Chain, error error) {
	res := gdb.DB.WithContext(ctx).Model(&model.Chain{}).Find(&chains)
	if res.Error != nil {
		return nil, res.Error
	}
	return chains, nil
}

func (l *ChainLogic) Save(ctx context.Context, chainModel *model.Chain) error {
	return gdb.DB.WithContext(ctx).Model(&model.Chain{}).Save(&chainModel).Error
}
