# points
```text
需求：
1、部署一个带mint和burn功能的erc20合约，铸造销毁几个token，转移几个token，来构造事件
2、使用go语言写一个后端服务来追踪合约事件，重建用户的余额
3、以太坊延迟六个区块，确保区块链不会回滚
4、加上积分计算功能，起一个定时任务，每小时根据用户的余额来计算用户的积分，暂定积分是余额*0.05
5、要记录用户的所有余额变化，根据这个来计算积分，这样更准确一些
6、需要维护一下用户的总余额表以及总积分表，还有一个用户的余额变动记录表
7、需要支持多链逻辑，比如支持sepolia， base sepolia
举个例子：
用户在15：00的时候0个token，15：10分有100个token，15：30有200个token
计算积分的时候，需要考虑用户的余额变化

比如此时是16：00启动定时任务了来计算积分，应该是100*0.05*20/60+200*0.05*30/60
考虑一个场景，如果程序错误了，或者rpc有问题，导致好几天没有计算积分。此时应该如何正确回溯？
1、可以考虑引入中间件来计算积分
2、可以考虑引入回滚机制与临时缓存来做分叉处理
特别注意积分计算的可中断性，与可恢复性
```

## 项目说明
```sql
--表结构设计
--链表
CREATE TABLE chain (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER NOT NULL COMMENT '链ID', 
    name VARCHAR(255) NOT NULL COMMENT '合约名称', 
    symbol VARCHAR(10) NOT NULL COMMENT '合约符号',
    rpc_url VARCHAR(255) NOT NULL COMMENT 'RPC地址',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否活跃',
    start_block INTEGER NOT NULL DEFAULT 0 COMMENT '开始区块高度(开始追踪的区块高度)',
    last_confirmed_block BIGINT NOT NULL DEFAULT 0 COMMENT '最后确认的区块高度',
    last_proccesed_block INTEGER NOT NULL DEFAULT 0 COMMENT '最后处理的区块高度', 
    last_calculated_block INTEGER NOT NULL DEFAULT 0 COMMENT '最后计算积分的区块高度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--用户表
CREATE TABLE user (
    id SERIAL PRIMARY KEY,
    user_address VARCHAR(42) PRIMARY KEY,
    status INTEGER NOT NULL DEFAULT 1 COMMENT '用户状态 1:正常 0:禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--用户余额表
CREATE TABLE user_balance (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL COMMENT '用户ID', 
    chain_id INTEGER NOT NULL COMMENT '链ID',
    balance INTEGER NOT NULL COMMENT '余额',
    last_updated_block INTEGER NOT NULL COMMENT '已更新区块高度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--用户余额变动记录表
CREATE TABLE user_balance_change (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL COMMENT '用户ID', 
    chain_id INTEGER NOT NULL COMMENT '链ID',
    change_type VARCHAR(20) NOT NULL COMMENT '变动类型mint、burn、transfer_in、transfer_out', 
    amount DECIMAL(78,0) NOT NULL COMMENT '变动金额正数表示增加，负数表示减少', 
    balance_after DECIMAL(78,0) NOT NULL COMMENT '变动后的余额',
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
    block_time BIGINT NOT NULL COMMENT '区块时间',
    block_height INTEGER NOT NULL COMMENT '区块高度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--用户积分表
CREATE TABLE user_points (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL COMMENT '用户ID', 
    chain_id INTEGER NOT NULL COMMENT '链ID',
    points INTEGER NOT NULL COMMENT '积分',
    last_updated_block INTEGER NOT NULL COMMENT '已更新区块高度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--用户积分变动记录表
CREATE TABLE user_points_change (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL COMMENT '用户ID', 
    chain_id INTEGER NOT NULL COMMENT '链ID',
    from_block_height INTEGER NOT NULL COMMENT '变动前区块高度',
    to_block_height INTEGER NOT NULL COMMENT '变动后区块高度',
    from_block_time BIGINT NOT NULL COMMENT '变动前区块时间',
    to_block_time BIGINT NOT NULL COMMENT '变动后区块时间',
    calculated_points JSONB NOT NULL COMMENT '计算出的积分JSON',
    status INTEGER NOT NULL DEFAULT 1 COMMENT '状态 1:成功 0:失败',
    reason VARCHAR(255) COMMENT '失败原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
## 功能说明
```text
1. 链余额追踪服务
    - 负责监听每个链的新区块确认事件
    - 当新的区块被确认时，更新链表的最后确认区块高度，插入用户余额变动记录，更新用户余额表
    - 用户不存在时，插入用户记录
    - 提供查询接口，用于查看链追踪信息，包括最后确认区块高度、最后处理区块高度、最后计算积分区块高度
    - 使用线程池循环查询每个链的新区块确认事件
    - 回滚机制与临时缓存
        - 用户余额变动事件缓存至redis
        - 延后N个区块高度处理，确保用户余额变动事件被处理
        - 当发生分叉或回滚时，删除延后的N个缓存区块数据，更新链表的最后确认区块高度
2. 积分计算服务
    - 定时任务触发，从链读取已计算区块高度
    - 从用户余额变动记录表中读取该区块高度以后的余额变动，分批处理(每次处理N条)
    - 根据变动计算用户的积分变化
    - 更新用户积分表和积分变动记录表
3. 计算可中断性与可恢复性
    - 如果积分计算过程中出现错误，记录错误信息到积分变动记录表
    - 提供查询接口，用于查看失败的积分计算记录
    - 提供重试机制，允许手动触发失败记录的积分计算
    - 可以从最后计算的区块高度继续计算，实现可恢复性
    - 基于链状态及用户状态可控制是否计算整个链或用户的积分
4、技术实现
    - 使用go语言实现
    - 数据库使用mysql
    - 缓存使用redis
    - 线程池使用go语言的sync.WaitGroup实现
    - 定时任务使用go语言的time.Ticker实现(若是集群服务分布式调度用啥？)
    - 提供配置文件，用于配置链ID、区块确认延迟高度、批量处理数量、定时任务间隔等参数
    - 使用go语言的context包实现任务取消与超时控制
    - 使用gorm管理db、gin框架实现http接口
```    