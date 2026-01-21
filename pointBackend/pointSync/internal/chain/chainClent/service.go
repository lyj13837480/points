package chain

import (
	"context"
	"errors"
	"math/big"
	"pointSync/internal/chain"

	"pointSync/internal/chain/chainClent/evmclient"
	logTypes "pointSync/internal/chain/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type ChainClient interface {
	FilterLogs(ctx context.Context, q logTypes.FilterQuery) ([]interface{}, error)
	BlockTimeByNumber(context.Context, *big.Int) (uint64, error)
	Client() interface{}
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CallContractByChain(ctx context.Context, param logTypes.CallParam) (interface{}, error)
	BlockNumber() (uint64, error)
	BlockWithTxs(ctx context.Context, blockNumber uint64) (interface{}, error)
	BlockWithHash(ctx context.Context, blockNumber uint64) (common.Hash, error)
}

func New(chainID int64, nodeUrl string) (ChainClient, error) {
	switch chainID { //多链支持
	case chain.EthChainID, chain.OptimismChainID, chain.SepoliaChainID:
		return evmclient.New(nodeUrl)
	default:
		return nil, errors.New("unsupported chain id")
	}
}
