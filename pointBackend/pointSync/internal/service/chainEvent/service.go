package chainevent

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	chain "pointSync/internal/chain/chainClent"
	"pointSync/internal/chain/types"
	"pointSync/internal/logger/xzap"
	"pointSync/internal/logic"
	"pointSync/internal/model"
	"pointSync/internal/stores/xkv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/threading"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	SleepDuration = 5                                            //秒
	SleepDB       = 5                                            //秒
	BlockBatch    = 10                                           //每次处理区块数量
	XkvExpire     = 10 * 60 * 60                                 //xkv过期时间
	BlockBuffer   = 5                                            //区块缓冲区间
	AddressZero   = "0x0000000000000000000000000000000000000000" //地址0
	HexPrefix     = "0x"
	//合约abi 使用一行文本
	contractAbi = `[{"type":"constructor","inputs":[{"name":"name","type":"string","internalType":"string"},{"name":"symbol","type":"string","internalType":"string"},{"name":"initialOwner","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"ADMIN","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"allowance","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"spender","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"approve","inputs":[{"name":"spender","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"balanceOf","inputs":[{"name":"account","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"burn","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"decimals","inputs":[],"outputs":[{"name":"","type":"uint8","internalType":"uint8"}],"stateMutability":"view"},{"type":"function","name":"mint","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"amount","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"name","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"owner","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"pause","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"paused","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"renounceOwnership","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"symbol","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"view"},{"type":"function","name":"totalSupply","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transferFrom","inputs":[{"name":"from","type":"address","internalType":"address"},{"name":"to","type":"address","internalType":"address"},{"name":"value","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"nonpayable"},{"type":"function","name":"transferOwnership","inputs":[{"name":"newOwner","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"unpause","inputs":[],"outputs":[],"stateMutability":"nonpayable"},{"type":"event","name":"Approval","inputs":[{"name":"owner","type":"address","indexed":true,"internalType":"address"},{"name":"spender","type":"address","indexed":true,"internalType":"address"},{"name":"value","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"BURN","inputs":[{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"balance","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"MINT","inputs":[{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"balance","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"OwnershipTransferred","inputs":[{"name":"previousOwner","type":"address","indexed":true,"internalType":"address"},{"name":"newOwner","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"Paused","inputs":[{"name":"account","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"event","name":"TRANSFERFROM","inputs":[{"name":"from","type":"address","indexed":true,"internalType":"address"},{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"fromBalance","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"toBalance","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"Transfer","inputs":[{"name":"from","type":"address","indexed":true,"internalType":"address"},{"name":"to","type":"address","indexed":true,"internalType":"address"},{"name":"value","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"Unpaused","inputs":[{"name":"account","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"error","name":"ERC20InsufficientAllowance","inputs":[{"name":"spender","type":"address","internalType":"address"},{"name":"allowance","type":"uint256","internalType":"uint256"},{"name":"needed","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ERC20InsufficientBalance","inputs":[{"name":"sender","type":"address","internalType":"address"},{"name":"balance","type":"uint256","internalType":"uint256"},{"name":"needed","type":"uint256","internalType":"uint256"}]},{"type":"error","name":"ERC20InvalidApprover","inputs":[{"name":"approver","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidReceiver","inputs":[{"name":"receiver","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidSender","inputs":[{"name":"sender","type":"address","internalType":"address"}]},{"type":"error","name":"ERC20InvalidSpender","inputs":[{"name":"spender","type":"address","internalType":"address"}]},{"type":"error","name":"EnforcedPause","inputs":[]},{"type":"error","name":"ExpectedPause","inputs":[]},{"type":"error","name":"OwnableInvalidOwner","inputs":[{"name":"owner","type":"address","internalType":"address"}]},{"type":"error","name":"OwnableUnauthorizedAccount","inputs":[{"name":"account","type":"address","internalType":"address"}]}]`
)

var waitBlock = map[string]uint64{
	"ETH": 10,
}

var _chainIDFuncMap = map[int64]map[string]string{
	11155111: {
		"0xa30bd593763e568665ca0917c6ad910a757ccff08086ca4d1574888caa9e482b": "MINT",
		"0x9c89a2d315b61ccb5c7e40f7fe8fe63317b1079d3321471cbf4681273b5c1fb8": "BURN",
		"0xbf7c5370684103b9b5021af3a89f3cf8741b2de7312c598f7ac91af709d8c9b1": "TRANSFER",
	},
}

type Service struct {
	ChainClient chain.ChainClient
	Ctx         context.Context
	Db          *gorm.DB
	Xkv         *xkv.Store
	Chain       *model.Chain
	ParsedAbi   abi.ABI
}

type BlockEvent struct {
	Hash common.Hash
	Logs []ethereumTypes.Log
}

func New(db *gorm.DB, chainClient chain.ChainClient, ctx context.Context, chain *model.Chain, xkv *xkv.Store) *Service {
	parsedAbi, _ := abi.JSON(strings.NewReader(contractAbi)) //
	return &Service{
		Db:          db,
		ChainClient: chainClient,
		Ctx:         ctx,
		Chain:       chain,
		ParsedAbi:   parsedAbi,
		Xkv:         xkv,
	}
}

func (s *Service) Start() {
	threading.GoSafe(s.balanceEventKv) // 监听合约事件，将事件保存到xkv
	threading.GoSafe(s.balanceEventDb) // 从xkv中读取事件，保存到db
}

// balanceEventKv 监听合约事件，将事件保存到xkv
func (s *Service) balanceEventKv() {
	syncBlock := s.Chain.LastConfirmedBlock
	xzap.WithContext(s.Ctx).Info("balanceEventKv start")
	for {
		select {
		case <-s.Ctx.Done():
			log.Info("balanceEvent ctx done")
			return
		default:
		}
		currentBlock, err := s.ChainClient.BlockNumber()
		if err != nil {
			log.Error("balanceEvent get current block error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		if syncBlock > currentBlock-waitBlock[s.Chain.Name] {
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
			Addresses: []string{s.Chain.ContractAddress},
		}

		logs, err := s.ChainClient.FilterLogs(s.Ctx, filterQuery)
		if err != nil {
			log.Error("balanceEvent filter logs error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
		waitSaveBlock := map[uint64]BlockEvent{}
		// 先获取所有区块的hash
		for i := syncBlock; i <= endBlock; i++ {
			blockEvent := BlockEvent{}
			blockHash, err := s.ChainClient.BlockWithHash(s.Ctx, i)
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
			s.Xkv.Write(fmt.Sprintf("%d", blockNum), blockEvent, int(XkvExpire))
		}
		syncBlock = endBlock + 1
		s.Chain.LastConfirmedBlock = syncBlock
		if err := s.Db.Model(&model.Chain{}).Where("chain_id = ?", s.Chain.ChainID).Updates(map[string]interface{}{
			"last_confirmed_block": syncBlock,
		}).Error; err != nil {
			log.Error("balanceEvent save chain error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
	}
}

// balanceEventDb 读取缓存合约事件，将事件保存到数据库
func (s *Service) balanceEventDb() {

	for {
		select {
		case <-s.Ctx.Done():
			log.Info("balanceEventDb ctx done")
			return
		default:
		}
		confirmedBlock := s.Chain.LastConfirmedBlock
		processedBlock := s.Chain.LastProcessedBlock
		if processedBlock >= confirmedBlock-BlockBuffer {
			time.Sleep(SleepDB * time.Second)
			continue
		}
		var blockEvent BlockEvent
		ok, err := s.Xkv.Read(fmt.Sprintf("%d", processedBlock), &blockEvent)
		if err != nil {
			log.Error("balanceEventDb read xkv error", "err", err)
			time.Sleep(SleepDB * time.Second)
			continue
		}
		if !ok {
			time.Sleep(SleepDB * time.Second)
			continue
		}

		blockHash, err := s.ChainClient.BlockWithHash(s.Ctx, processedBlock)
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
			switch _chainIDFuncMap[s.Chain.ChainID][ethLog.Topics[0].Hex()] {
			case "MINT":
				s.mint(&ethLog)
			case "BURN":
				s.burn(&ethLog)
			case "TRANSFER":
				s.transfer(&ethLog)
			default:
				xzap.WithContext(s.Ctx).Warn("unknown event", zap.String("topic", ethLog.Topics[0].Hex()))
			}
		}
		// 更新链数据
		processedBlock++
		s.Chain.LastProcessedBlock = processedBlock
		if err := s.Db.Model(&model.Chain{}).Where("chain_id = ?", s.Chain.ChainID).Updates(map[string]interface{}{
			"last_processed_block": processedBlock,
		}).Error; err != nil {
			log.Error("balanceEvent save chain error", "err", err)
			time.Sleep(SleepDuration * time.Second)
			continue
		}
	}
}

func (s *Service) mint(log *ethereumTypes.Log) {
	var event struct {
		To      common.Address
		Amount  *big.Int
		Balance *big.Int
	}
	// 解析事件参数
	if err := s.ParsedAbi.UnpackIntoInterface(&event, "MINT", log.Data); err != nil {
		xzap.WithContext(s.Ctx).Error("mint unpack log data error", zap.Error(err))
	}
	event.To = common.HexToAddress(log.Topics[1].Hex())
	user, err := logic.UserLogicInstance.GetUserByAddress(s.Ctx, event.To.String())
	s.Db.Transaction(func(tx *gorm.DB) error {
		if user.ID == 0 {
			user.UserAddress = event.To.String()
			err = tx.WithContext(s.Ctx).Model(&model.User{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&user).Error
			if err != nil {
				xzap.WithContext(s.Ctx).Error("mint create user error", zap.Error(err))
				return err
			}
			err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&model.UserBalance{
				UserID:           user.ID,
				ChainID:          s.Chain.ChainID,
				Balance:          decimal.NewFromBigInt(event.Amount, 0),
				LastUpdatedBlock: int64(log.BlockNumber),
			}).Error

			if err != nil {
				xzap.WithContext(s.Ctx).Error("mint create user balance error", zap.Error(err))
				return err
			}
		}
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint get user error", zap.Error(err))
			return err
		}
		change := &model.UserBalanceChange{
			UserID:       user.ID, //TODO 根据日志解析用户ID
			ChainID:      s.Chain.ChainID,
			ChangeType:   _chainIDFuncMap[s.Chain.ChainID][log.Topics[0].String()],
			Amount:       decimal.NewFromBigInt(event.Amount, 0), //TODO 根据日志解析变动金额
			TxHash:       log.TxHash.Hex(),
			BlockTime:    int64(log.BlockTimestamp),
			BlockHeight:  int64(log.BlockNumber),
			BalanceAfter: decimal.NewFromBigInt(event.Balance, 0), //TODO 根据日志解析变动金额
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalanceChange{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(change).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint save user balance change error", zap.Error(err))
			return err
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", user.ID, int(s.Chain.ChainID)).Updates(map[string]interface{}{
			"balance":            decimal.NewFromBigInt(event.Balance, 0),
			"last_updated_block": int64(log.BlockNumber),
		}).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint save user balance error", zap.Error(err))
			return err
		}
		return nil
	})
}

func (s *Service) burn(log *ethereumTypes.Log) {
	// TODO 实现burn事件处理逻辑
	var event struct {
		To      common.Address
		Amount  *big.Int
		Balance *big.Int
	}
	// 解析事件参数
	if err := s.ParsedAbi.UnpackIntoInterface(&event, "BURN", log.Data); err != nil {
		xzap.WithContext(s.Ctx).Error("burn unpack log data error", zap.Error(err))
	}
	event.To = common.HexToAddress(log.Topics[1].Hex())
	user, err := logic.UserLogicInstance.GetUserByAddress(s.Ctx, event.To.String())
	s.Db.Transaction(func(tx *gorm.DB) error {
		if user.ID == 0 {
			user.UserAddress = event.To.String()
			err = tx.WithContext(s.Ctx).Model(&model.User{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&user).Error
			if err != nil {
				xzap.WithContext(s.Ctx).Error("mint create user error", zap.Error(err))
				return err
			}
			err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&model.UserBalance{
				UserID:           user.ID,
				ChainID:          s.Chain.ChainID,
				Balance:          decimal.NewFromBigInt(event.Amount, 0),
				LastUpdatedBlock: int64(log.BlockNumber),
			}).Error

			if err != nil {
				xzap.WithContext(s.Ctx).Error("mint create user balance error", zap.Error(err))
				return err
			}
		}
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint get user error", zap.Error(err))
			return err
		}
		change := &model.UserBalanceChange{
			UserID:       user.ID, //TODO 根据日志解析用户ID
			ChainID:      s.Chain.ChainID,
			ChangeType:   _chainIDFuncMap[s.Chain.ChainID][log.Topics[0].String()],
			Amount:       decimal.NewFromBigInt(event.Amount, 0).Neg(), // 根据日志解析变动金额 负数表示减少
			TxHash:       log.TxHash.Hex(),
			BlockTime:    int64(log.BlockTimestamp),
			BlockHeight:  int64(log.BlockNumber),
			BalanceAfter: decimal.NewFromBigInt(event.Balance, 0), //TODO 根据日志解析变动金额
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalanceChange{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(change).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint save user balance change error", zap.Error(err))
			return err
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", user.ID, int(s.Chain.ChainID)).Updates(map[string]interface{}{
			"balance":            decimal.NewFromBigInt(event.Balance, 0),
			"last_updated_block": int64(log.BlockNumber),
		}).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("mint save user balance error", zap.Error(err))
			return err
		}
		return nil
	})
}

func (s *Service) transfer(log *ethereumTypes.Log) {
	// TODO 实现transferFrom事件处理逻辑
	var event struct {
		From        common.Address
		To          common.Address
		Amount      *big.Int
		FromBalance *big.Int
		ToBalance   *big.Int
	}
	// 解析事件参数
	if err := s.ParsedAbi.UnpackIntoInterface(&event, "TRANSFERFROM", log.Data); err != nil {
		xzap.WithContext(s.Ctx).Error("transfer unpack log data error", zap.Error(err))
	}
	// // 过滤地址0 若地址为0说明是创建或销毁
	// if event.From.String() == AddressZero || event.To.String() == AddressZero {
	// 	return
	// }
	// 处理转账逻辑
	// 处理发送方
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.To = common.HexToAddress(log.Topics[2].Hex())
	user, err := logic.UserLogicInstance.GetUserByAddress(s.Ctx, event.From.String())
	s.Db.Transaction(func(tx *gorm.DB) error {
		if user.ID == 0 {
			user.UserAddress = event.From.String()
			err = tx.WithContext(s.Ctx).Model(&model.User{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&user).Error
			if err != nil {
				xzap.WithContext(s.Ctx).Error("transferOut create user error", zap.Error(err))
				return err
			}
			err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&model.UserBalance{
				UserID:           user.ID,
				ChainID:          s.Chain.ChainID,
				Balance:          decimal.NewFromBigInt(event.Amount, 0),
				LastUpdatedBlock: int64(log.BlockNumber),
			}).Error

			if err != nil {
				xzap.WithContext(s.Ctx).Error("transferOut create user balance error", zap.Error(err))
				return err
			}
		}
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferOut get user error", zap.Error(err))
			return err
		}
		change := &model.UserBalanceChange{
			UserID:       user.ID, //TODO 根据日志解析用户ID
			ChainID:      s.Chain.ChainID,
			ChangeType:   "TRANSFER_OUT",
			Amount:       decimal.NewFromBigInt(event.Amount, 0).Neg(), // 根据日志解析变动金额 负数表示减少
			TxHash:       log.TxHash.Hex(),
			BlockTime:    int64(log.BlockTimestamp),
			BlockHeight:  int64(log.BlockNumber),
			BalanceAfter: decimal.NewFromBigInt(event.FromBalance, 0),
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalanceChange{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(change).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferOut save user balance change error", zap.Error(err))
			return err
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", user.ID, int(s.Chain.ChainID)).Updates(map[string]interface{}{
			"balance":            decimal.NewFromBigInt(event.FromBalance, 0),
			"last_updated_block": int64(log.BlockNumber),
		}).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferOut save user balance error", zap.Error(err))
			return err
		}
		return nil
	})
	// 处理接收方
	user, err = logic.UserLogicInstance.GetUserByAddress(s.Ctx, event.To.String())
	s.Db.Transaction(func(tx *gorm.DB) error {
		if user.ID == 0 {
			user.UserAddress = event.To.String()
			err = tx.WithContext(s.Ctx).Model(&model.User{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&user).Error
			if err != nil {
				xzap.WithContext(s.Ctx).Error("transferIn create user error", zap.Error(err))
				return err
			}
			err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&model.UserBalance{
				UserID:           user.ID,
				ChainID:          s.Chain.ChainID,
				Balance:          decimal.NewFromBigInt(event.Amount, 0),
				LastUpdatedBlock: int64(log.BlockNumber),
			}).Error

			if err != nil {
				xzap.WithContext(s.Ctx).Error("transferIn create user balance error", zap.Error(err))
				return err
			}
		}
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferIn get user error", zap.Error(err))
			return err
		}
		change := &model.UserBalanceChange{
			UserID:       user.ID, //TODO 根据日志解析用户ID
			ChainID:      s.Chain.ChainID,
			ChangeType:   "TRANSFER_IN",
			Amount:       decimal.NewFromBigInt(event.Amount, 0), // 根据日志解析变动金额
			TxHash:       log.TxHash.Hex(),
			BlockTime:    int64(log.BlockTimestamp),
			BlockHeight:  int64(log.BlockNumber),
			BalanceAfter: decimal.NewFromBigInt(event.ToBalance, 0), //TODO 根据日志解析变动金额
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalanceChange{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(change).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferIn save user balance change error", zap.Error(err))
			return err
		}
		var userBalance model.UserBalance
		err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", user.ID, int(s.Chain.ChainID)).First(&userBalance).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferIn get user balance error", zap.Error(err))
			return err
		}
		err = tx.WithContext(s.Ctx).Model(&model.UserBalance{}).Where("user_id = ? AND chain_id = ?", user.ID, int(s.Chain.ChainID)).Updates(map[string]interface{}{
			"balance":            decimal.NewFromBigInt(event.ToBalance, 0),
			"last_updated_block": int64(log.BlockNumber),
		}).Error
		if err != nil {
			xzap.WithContext(s.Ctx).Error("transferIn save user balance error", zap.Error(err))
			return err
		}
		return nil
	})

}

// 处理分叉情况,hash不一致,需要回滚，删除已confirmedBlock之后的数据
func (s *Service) handleFork(blockNumber uint64) {
	// TODO 实现handleFork事件处理逻辑
	// 区块回退至不一致的高度
	// 更新链数据
	s.Chain.LastConfirmedBlock = blockNumber - 1
	s.Chain.LastProcessedBlock = blockNumber - 1
	if s.Chain.LastCalculatedBlock >= blockNumber {
		s.Chain.LastCalculatedBlock = blockNumber - 1
	}
	// 保存更新后的链状态
	err := logic.ChainLogicInstance.Save(s.Ctx, s.Chain)
	if err != nil {
		xzap.WithContext(s.Ctx).Error("handleFork save chain error", zap.Error(err))
	}

}
