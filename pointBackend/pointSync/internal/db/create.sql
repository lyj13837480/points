

--链表
CREATE TABLE chain (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键',
    chain_id INTEGER NOT NULL COMMENT '链ID', 
    name VARCHAR(255) NOT NULL COMMENT '合约名称', 
    symbol VARCHAR(10) NOT NULL COMMENT '合约符号',
    rpc_url VARCHAR(255) NOT NULL COMMENT 'RPC地址',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否活跃',
    start_block BIGINT NOT NULL DEFAULT 0 COMMENT '开始区块高度',
    last_confirmed_block BIGINT NOT NULL DEFAULT 0 COMMENT '最后确认的区块高度',
    last_processed_block BIGINT NOT NULL DEFAULT 0 COMMENT '最后处理的区块高度', 
    last_calculated_block BIGINT NOT NULL DEFAULT 0 COMMENT '最后计算积分的区块高度',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY idx_chain_id (chain_id),  -- 或使用 INDEX 关键字
    UNIQUE KEY uk_contract_chain (contract_address, chain_id),
    KEY idx_symbol (symbol),
    KEY idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区块链配置表';
-- 用户表（修正：主键冲突、SERIAL语法）
CREATE TABLE user (
id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
user_address VARCHAR(42) NOT NULL COMMENT '用户地址',
status INTEGER NOT NULL DEFAULT 1 COMMENT '用户状态 1:正常 0:禁用',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
UNIQUE KEY uk_user_address (user_address),
KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
-- 用户余额表（修正：SERIAL语法，添加索引）
CREATE TABLE user_balance (
id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
user_id INTEGER NOT NULL COMMENT '用户ID',
chain_id INTEGER NOT NULL COMMENT '链ID',
balance INTEGER NOT NULL COMMENT '余额',
last_updated_block BIGINT NOT NULL DEFAULT 0 COMMENT '已更新区块高度',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
UNIQUE KEY uk_user_chain (user_id, chain_id),
KEY idx_user_id (user_id),
KEY idx_chain_id (chain_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户余额表';
-- 用户余额变动记录表（修正：SERIAL语法、DECIMAL精度、区块高度类型、添加索引）
CREATE TABLE user_balance_change (
id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
user_id INTEGER NOT NULL COMMENT '用户ID',
chain_id INTEGER NOT NULL COMMENT '链ID',
change_type VARCHAR(20) NOT NULL COMMENT '变动类型mint、burn、transfer_in、transfer_out',
amount DECIMAL(65,0) NOT NULL COMMENT '变动金额正数表示增加，负数表示减少',
balance_after DECIMAL(65,0) NOT NULL COMMENT '变动后的余额',
tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
block_time BIGINT NOT NULL COMMENT '区块时间',
block_height BIGINT NOT NULL COMMENT '区块高度',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
KEY idx_user_id (user_id),
KEY idx_chain_id (chain_id),
KEY idx_block_height (block_height),
KEY idx_tx_hash (tx_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户余额变动记录表';
-- 用户积分表（修正：SERIAL语法、区块高度类型，添加索引）
CREATE TABLE user_points (
id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
user_id INTEGER NOT NULL COMMENT '用户ID',
chain_id INTEGER NOT NULL COMMENT '链ID',
points INTEGER NOT NULL COMMENT '积分',
last_updated_block BIGINT NOT NULL DEFAULT 0 COMMENT '已更新区块高度',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
UNIQUE KEY uk_user_chain (user_id, chain_id),
KEY idx_user_id (user_id),
KEY idx_chain_id (chain_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户积分表';
-- 用户积分变动记录表（修正：SERIAL语法、JSONB→JSON、区块高度类型，添加索引）
CREATE TABLE user_points_change (
id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
user_id INTEGER NOT NULL COMMENT '用户ID',
chain_id INTEGER NOT NULL COMMENT '链ID',
from_block_height BIGINT NOT NULL COMMENT '变动前区块高度',
to_block_height BIGINT NOT NULL COMMENT '变动后区块高度',
from_block_time BIGINT NOT NULL COMMENT '变动前区块时间',
to_block_time BIGINT NOT NULL COMMENT '变动后区块时间',
calculated_points JSON NOT NULL COMMENT '计算出的积分JSON',
status INTEGER NOT NULL DEFAULT 1 COMMENT '状态 1:成功 0:失败',
reason VARCHAR(255) COMMENT '失败原因',
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
KEY idx_user_id (user_id),
KEY idx_chain_id (chain_id),
KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户积分变动记录表';