package gdb

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	timeUtil "pointSync/internal/util/time"
)

var DB *gorm.DB

type Config struct {
	User               string `json:"user"`                                                       // 用户
	Password           string ` json:"password"`                                                  // 密码
	Host               string `json:"host"`                                                       // 地址
	Port               int    `json:"port"`                                                       // 端口
	Database           string `json:"database"`                                                   // 数据库
	MaxIdleConns       int    `mapstructure:"max_idle_conns" json:"max_idle_conns"`               // 最大空闲连接数
	MaxOpenConns       int    `mapstructure:"max_open_conns" json:"max_open_conns"`               // 最大打开连接数
	MaxConnMaxLifetime int64  `mapstructure:"max_conn_max_lifetime" json:"max_conn_max_lifetime"` // 连接复用时间
	LogLevel           string `mapstructure:"log_level" json:"log_level"`                         // 日志级别，枚举（info、warn、error和silent）
}

func (c *Config) GetDataSource() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}

// GetMySQLConfig 获取GORM MySQL相关配置
func (c *Config) GetMySQLConfig() mysql.Config {
	return mysql.Config{
		DSN:                       c.GetDataSource(),
		DefaultStringSize:         255,  // string类型字段默认长度
		DisableDatetimePrecision:  true, // 禁用datetime精度
		DontSupportRenameIndex:    true, // 禁用重命名索引
		DontSupportRenameColumn:   true, // 禁用重命名列名
		SkipInitializeWithVersion: true, // 禁用根据当前mysql版本自动配置
	}
}

// GetGormConfig 获取GORM相关配置
func (c *Config) GetGormConfig() *gorm.Config {
	gc := &gorm.Config{
		QueryFields: true, // 根据字段名称查询
		PrepareStmt: true, // 缓存预编译语句
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 数据表名单数
		},
		NowFunc: func() time.Time {
			return timeUtil.Now() // 当前时间载入时区
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用自动创建外键约束
	}

	logLevel := logger.Warn
	switch c.LogLevel {
	case "info":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	case "silent":
		logLevel = logger.Silent
	}

	gc.Logger = logger.Default.LogMode(logLevel)
	//gc.Logger = NewLogger(logLevel, 200*time.Millisecond) // 设置日志记录器

	return gc
}

func NewDB(c *Config) (*gorm.DB, error) {
	if c == nil {
		return nil, errors.New("gdb: illegal gdb configure")
	}

	db, err := gorm.Open(mysql.New(c.GetMySQLConfig()), c.GetGormConfig())
	if err != nil {
		return nil, errors.WithMessage(err, "gdb: open database connection err")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithMessage(err, "gdb: get database instance err")
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	if c.MaxConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(c.MaxConnMaxLifetime))
	}
	return db, nil
}

func MustNewDB(c *Config) {
	var err error
	DB, err = NewDB(c)
	if err != nil {
		panic(err)
	}
	sqlDB, _ := DB.DB()
	if pingErr := sqlDB.Ping(); pingErr != nil { //测试数据库连接
		fmt.Errorf("数据库连接失败:%s \n", pingErr)
		panic(pingErr)
	}
	fmt.Println("数据库连接成功 \n")
}
