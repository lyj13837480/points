package service

import (
	"context"
	pChain "pointSync/internal/chain/chainClent"
	"pointSync/internal/config"
	"pointSync/internal/logger/xzap"
	"pointSync/internal/logic"
	chainevent "pointSync/internal/service/chainEvent"
	"pointSync/internal/stores/gdb"
	"pointSync/internal/stores/xkv"

	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/zap"
)

type ServiceContext struct {
	ctx        context.Context
	chainevent *[]chainevent.Service
}

func New(ctx context.Context, cfg *config.Config) *ServiceContext {
	chains, err := logic.ChainLogicInstance.QueryChainData(ctx)
	if err != nil {
		log.Error("QueryChainData failed, err: %v", err)
	}
	chaineventServices := make([]chainevent.Service, 0, len(chains))
	for _, c := range chains {
		var chainClient pChain.ChainClient
		chainClient, err = pChain.New(c.ChainID, cfg.AnkrCfg.HttpsURL+cfg.AnkrCfg.ApiKey)
		if err != nil {
			log.Error("New chain client failed, err: %v", err)
		}
		xkvStore := xkv.NewStore(cfg) // 初始化键值存储
		chaineventServices = append(chaineventServices, *chainevent.New(gdb.DB, chainClient, ctx, &c, xkvStore))
	}
	return &ServiceContext{
		ctx:        ctx,
		chainevent: &chaineventServices,
	}
}

func (s *ServiceContext) Start() {
	for _, c := range *s.chainevent {
		c.Start()
		xzap.WithContext(c.Ctx).Info("chain event service started, chain: %s", zap.String("chain_name", c.Chain.Name))
	}
}
