package chainevent

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	chain "pointSync/pointSync/internal/chain/chainClent"
	"pointSync/pointSync/internal/chain/types"
	"pointSync/pointSync/internal/logger/xzap"
	"pointSync/pointSync/internal/logic"
	"pointSync/pointSync/internal/model"
	"pointSync/pointSync/internal/stores/xkv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

const (
	SleepDuration = 10                         //秒
	SleepDB       = 10                         //秒
	BlockBatch    = 10                         //每次处理区块数量
	XkvExpire     = 24 * 60 * 60 * time.Second //xkv过期时间
	BlockBuffer   = 5                          //区块缓冲区间
	contractAbi   = `[{"type":"constructor","inputs":[{"name":"name","type":"string","internalType":"string"},{"name":"symbol","type":"string","internalType":"string"},{"name":"initialOwner","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"ADMIN","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"allowance","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"spender","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"burn","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"decimals","inputs":[],"outputs":[{"name":"","type":"uint8","internalType":"uint8"}],"stateMutability":"view"},{"type":"function","name":"mint","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"name","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"owner","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"pause","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"paused","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"renounceOwnership","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"symbol","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"totalSupply","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transferFrom","inputs":[{"name":"from","type":"address","internalType":"address"},{"name":"to","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transferOwnership","inputs":[{"name":"newOwner","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"unpause","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"Approval","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"spender","type":"address","indexed":true,"internalType":"address"},{"name":"value","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"BURN","inputs":[{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"MINT","inputs":[{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"OwnershipTransferred","inputs":[{"name":"previousOwner","type":"address","indexed":true,"internalType":"address"},{"name":"newOwner","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"Paused","inputs":[{"name":"account","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"event","name":"Transfer","inputs":[{"name":"from","type":"address","indexed":true,"internalType":"address"},{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"value","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"Unpaused","inputs":[{"name":"account","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"error","name":"ERC20InsufficientAllowance","inputs":[{"name":"spender","type":"address","internalType":"address"},{"name":"allowance","type":"uint256","internalType":"uint256"},{"name":"needed","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ERC20InsufficientBalance","inputs":[{"name":"sender","type":"address","internalType":"address"},{"name":"balance","type":"uint256","internalType":"uint256"},{"name":"needed","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ERC20InvalidApprover","inputs":[{"name":"approver","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidReceiver","inputs":[{"name":"receiver","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidSender","inputs":[{"name":"sender","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidSpender","inputs":[{"name":"spender","type":"address","internalType":"address"}]},{"type":"error","name":"EnforcedPause","inputs":[]},{"type":"error","name":"ExpectedPause","inputs":[]},{"type":"error","name":"OwnableInvalidOwner","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"OwnableUnauthorizedAccount","inputs":[{"name":"account","type":"address","internalType":"address"}]}]`
)

var waitBlock = map[string]uint64{
	"ETH": 10,
}

var _chainIDFuncMap = map[int64]map[string]string{
	11155111: {
		"MINT":         "MINT",
		"BURN":         "BURN",
		"TRANSFER":     "TRANSFER",
		"TRANSFERFROM": "TRANSFERFROM",
	},
}

type Service struct {
	chainClient chain.ChainClient
	ctx         context.Context
	db          *gorm.DB
	xkv         *xkv.Store
	chain       model.Chain
	parsedAbi   abi.ABI
}

type BlockEvent struct {
	Hash common.Hash
	Logs []ethereumTypes.Log
}

func New(db *gorm.DB, chainClient chain.ChainClient, ctx context.Context, chain model.Chain) *Service {
	parsedAbi, _ := abi.JSON(strings.NewReader(contractAbi)) //
	return &Service{
		db:          db,
		chainClient: chainClient,
		ctx:         ctx,
		chain:       chain,
		parsedAbi:   parsedAbi,
	}
}

func (s *Service) Start() {
	threading.GoSafe(s.balanceEventKv)
}

// balanceEventKv 监听合约事件，将事件保存到xkv
func (s *Service) balanceEventKv() {
	syncBlock := s.chain.LastConfirmedBlock

	for {
		select {
		case <-s.ctx.Done():
			log.Info("balanceEvent ctx done")
			return
		default:
		}
		currentBlock, err := s.chainClient.BlockNumber()
		if err != nil {
			log.Error("balanceEvent get current block error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		if syncBlock > currentBlock-waitBlock[s.chain.Name] {
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		endBlock := syncBlock + BlockBatch
		if endBlock > currentBlock {
			endBlock = currentBlock
		}
		filterQuery := types.FilterQuery{
			FromBlock: big.NewInt(int64(syncBlock)),
			ToBlock:   big.NewInt(int64(endBlock)),
			Addresses: []string{s.chain.ContractAddress},
		}

		logs, err := s.chainClient.FilterLogs(s.ctx, filterQuery)
		if err != nil {
			log.Error("balanceEvent filter logs error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		waitSaveBlock := map[uint64]BlockEvent{}
		// 先获取所有区块的hash
		for i := syncBlock; i <= endBlock; i++ {
			blockEvent := BlockEvent{}
			blockHash, err := s.chainClient.BlockWithHash(s.ctx, i)
			if err != nil {
				log.Error("balanceEvent get block hash error", "err", err)
				time.Sleep(SleepDuration * time.Second)
				continue
			}
			blockEvent.Hash = blockHash
			waitSaveBlock[i] = blockEvent
		}
		// 分类日志到对应区块
		for _, vLog := range logs {
			ethLog := vLog.(ethereumTypes.Log)
			blockEvent := waitSaveBlock[ethLog.BlockNumber]
			blockEvent.Logs = append(blockEvent.Logs, ethLog)
			waitSaveBlock[ethLog.BlockNumber] = blockEvent
		}
		// 保存到xkv
		for blockNum, blockEvent := range waitSaveBlock {
			s.xkv.Write(fmt.Sprintf("%d", blockNum), blockEvent, int(XkvExpire.Seconds()))
		}
		syncBlock = endBlock + 1
		s.chain.LastConfirmedBlock = syncBlock
		if err := s.db.Model(&model.Chain{}).Where("chain_id = ?", s.chain.ChainID).Updates(map[string]interface{}{
			"last_confirmed_block": syncBlock,
		}).Error; err != nil {
			log.Error("balanceEvent save chain error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		time.Sleep(SleepDuration * time.Second)
	}
}

// balanceEventDb 读取缓存合约事件，将事件保存到数据库
func (s *Service) balanceEventDb() {

	for {
		select {
		case <-s.ctx.Done():
			log.Info("balanceEventDb ctx done")
			return
		default:
		}
		confirmedBlock := s.chain.LastConfirmedBlock
		processedBlock := s.chain.LastProcessedBlock
		if processedBlock >= confirmedBlock-BlockBuffer {
			time.Sleep(SleepDB * time.Second)
			continue
		}
		var blockEvent BlockEvent
		ok, err := s.xkv.Read(fmt.Sprintf("%d", processedBlock), &blockEvent)
		if err != nil {
			log.Error("balanceEventDb read xkv error", "err", err)
			time.Sleep(SleepDB * time.Second)
			continue
		}
		if !ok {
			time.Sleep(SleepDB * time.Second)
			continue
		}

		blockHash, err := s.chainClient.BlockWithHash(s.ctx, processedBlock)
		if err != nil {
			log.Error("balanceEvent get block hash error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		// 校验区块hash是否匹配
		if blockHash != blockEvent.Hash {
			log.Error("balanceEvent block hash not match", "blockHash", blockHash, "blockEvent.Hash", blockEvent.Hash)
			//TODO 处理分叉情况
			s.handleFork(processedBlock)
			time.Sleep(SleepDuration * time.Second)
			continue
		}

		for _, ethLog := range blockEvent.Logs {
			switch _chainIDFuncMap[s.chain.ChainID][ethLog.Topics[0].Hex()] {
			case "MINT":
				s.mint(&ethLog)
			case "BURN":
				s.burn(&ethLog)
			case "TRANSFERFROM":
				s.transferFrom(&ethLog)
			default:
				xzap.WithContext(s.ctx).Warn("unknown event", "topic", ethLog.Topics[0].Hex())
			}
		}
		// 更新链数据
		s.chain.LastProcessedBlock = processedBlock
		if err := logic.ChainLogicInstance.Save(s.ctx, &s.chain); err != nil {
			xzap.WithContext(s.ctx).Error("balanceEventDb save chain error", "err", err)
			time.Sleep(SleepDB * time.Second)
			continue
		}
		processedBlock++
	}
}

func (s *Service) mint(log *ethereumTypes.Log) {
	var event struct {
		From   common.Address
		To     common.Address
		Amount *big.Int
	}
	// 解析事件参数
	if err := s.parsedAbi.UnpackIntoInterface(&event, "Mint", log.Data); err != nil {
		xzap.WithContext(s.ctx).Error("mint unpack log data error", "err", err)

	}
	var user model.User
	user, err := logic.UserLogicInstance.GetUserByAddress(s.ctx, event.To.Hex())
	if err != nil {
		xzap.WithContext(s.ctx).Error("mint get user error", "err", err)

	}
	change := &model.UserBalanceChange{
		UserID:      user.ID, //TODO 根据日志解析用户ID
		ChainID:     int(s.chain.ChainID),
		ChangeType:  _chainIDFuncMap[s.chain.ChainID][log.Topics[0].Hex()],
		Amount:      event.Amount, //TODO 根据日志解析变动金额
		TxHash:      log.TxHash.Hex(),
		BlockTime:   int64(log.BlockTimestamp),
		BlockHeight: int64(log.BlockNumber),
	}

	err = logic.UserBalanceChangeLogicInstance.CreateUserBalanceChange(s.ctx, change)
	if err != nil {
		xzap.WithContext(s.ctx).Error("mint save user balance change error", "err", err)
	}
	err = logic.UserBalanceLogicInstance.Save(s.ctx, user.ID, s.chain.ChainID, event.Amount, int64(log.BlockNumber))
	if err != nil {
		xzap.WithContext(s.ctx).Error("mint save user balance error", "err", err)
	}
}

func (s *Service) burn(log *ethereumTypes.Log) {
	// TODO 实现burn事件处理逻辑
	var event struct {
		From   common.Address
		Amount *big.Int
	}
	// 解析事件参数
	if err := s.parsedAbi.UnpackIntoInterface(&event, "Burn", log.Data); err != nil {
		xzap.WithContext(s.ctx).Error("burn unpack log data error", "err", err)
	}
	var user model.User
	user, err := logic.UserLogicInstance.GetUserByAddress(s.ctx, event.From.Hex())
	if err != nil {
		xzap.WithContext(s.ctx).Error("burn get user error", "err", err)
	}
	change := &model.UserBalanceChange{
		UserID:      user.ID,
		ChainID:     int(s.chain.ChainID),
		ChangeType:  _chainIDFuncMap[s.chain.ChainID][log.Topics[0].Hex()],
		Amount:      event.Amount.Neg(event.Amount), //设置为负数
		TxHash:      log.TxHash.Hex(),
		BlockTime:   int64(log.BlockTimestamp),
		BlockHeight: int64(log.BlockNumber),
	}

	err = logic.UserBalanceChangeLogicInstance.CreateUserBalanceChange(s.ctx, change)
	if err != nil {
		xzap.WithContext(s.ctx).Error("burn save user balance change error", "err", err)
	}
	err = logic.UserBalanceLogicInstance.Save(s.ctx, user.ID, s.chain.ChainID, event.Amount.Neg(event.Amount), int64(log.BlockNumber))
	if err != nil {
		xzap.WithContext(s.ctx).Error("burn save user balance error", "err", err)
	}
}

func (s *Service) transferFrom(log *ethereumTypes.Log) {
	// TODO 实现transferFrom事件处理逻辑
	var event struct {
		From   common.Address
		To     common.Address
		Amount *big.Int
	}
	// 解析事件参数
	if err := s.parsedAbi.UnpackIntoInterface(&event, "TransferFrom", log.Data); err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom unpack log data error", "err", err)
	}
	// 处理转账逻辑
	// 处理发送方
	var user model.User
	user, err := logic.UserLogicInstance.GetUserByAddress(s.ctx, event.From.Hex())
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom get user error", "err", err)
	}
	change := &model.UserBalanceChange{
		UserID:      user.ID,
		ChainID:     int(s.chain.ChainID),
		ChangeType:  "TRANSFE_OUT",
		Amount:      event.Amount.Neg(event.Amount), //设置为负数
		TxHash:      log.TxHash.Hex(),
		BlockTime:   int64(log.BlockTimestamp),
		BlockHeight: int64(log.BlockNumber),
	}

	err = logic.UserBalanceChangeLogicInstance.CreateUserBalanceChange(s.ctx, change)
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom save user balance change error", "err", err)
	}

	err = logic.UserBalanceLogicInstance.Save(s.ctx, user.ID, s.chain.ChainID, event.Amount.Neg(event.Amount), int64(log.BlockNumber))
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom save user balance error", "err", err)
	}

	// 处理接收方
	user, err = logic.UserLogicInstance.GetUserByAddress(s.ctx, event.To.Hex())
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom get user error", "err", err)
	}
	change = &model.UserBalanceChange{
		UserID:      user.ID,
		ChainID:     int(s.chain.ChainID),
		ChangeType:  "TRANSFE_IN",
		Amount:      event.Amount, //设置为正数
		TxHash:      log.TxHash.Hex(),
		BlockTime:   int64(log.BlockTimestamp),
		BlockHeight: int64(log.BlockNumber),
	}

	err = logic.UserBalanceChangeLogicInstance.CreateUserBalanceChange(s.ctx, change)
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom save user balance change error", "err", err)
	}
	err = logic.UserBalanceLogicInstance.Save(s.ctx, user.ID, s.chain.ChainID, event.Amount, int64(log.BlockNumber))
	if err != nil {
		xzap.WithContext(s.ctx).Error("transferFrom save user balance error", "err", err)
	}
}

// 处理分叉情况,hash不一致,需要回滚，删除已confirmedBlock之后的数据
func (s *Service) handleFork(blockNumber uint64) {
	// TODO 实现handleFork事件处理逻辑
	confirmedBlock := s.chain.LastConfirmedBlock
	keys := []string{}
	for bn := blockNumber; bn <= confirmedBlock; bn++ {
		keys = append(keys, fmt.Sprintf("%d", bn))
	}
	// 删除xkv中的区块数据
	_, err := s.xkv.DelCtx(s.ctx, keys...)
	if err != nil {
		xzap.WithContext(s.ctx).Error("handleFork del block error", "err", err)
	}
	// 更新链数据
	s.chain.LastConfirmedBlock = blockNumber - 1
	s.chain.LastProcessedBlock = blockNumber - 1
	if s.chain.LastCalculatedBlock >= blockNumber {
		s.chain.LastCalculatedBlock = blockNumber - 1
	}
	// 保存更新后的链状态
	err = logic.ChainLogicInstance.Save(s.ctx, &s.chain)
	if err != nil {
		xzap.WithContext(s.ctx).Error("handleFork save chain error", "err", err)
	}

}
