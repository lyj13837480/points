// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"pointSync/pointSync/internal/config"
	"pointSync/pointSync/internal/handler"
	"pointSync/pointSync/internal/stores/gdb"
	"pointSync/pointSync/internal/stores/xkv"
	"pointSync/pointSync/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 获取程序所在目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("get executable path failed: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)

	// 构建配置文件的绝对路径
	absConfigPath := filepath.Join(execDir, *configFile)

	// 检查配置文件是否存在
	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		// 如果不存在，尝试使用相对路径
		absConfigPath = *configFile
	}

	var c config.Config
	conf.MustLoad(absConfigPath, &c)
	fmt.Printf("config %v\n", c)
	gdb.MustNewDB(&c.Mysql)                  // 初始化数据库连接
	xkv.NewStore(&c)                         // 初始化键值存储
	server := rest.MustNewServer(c.RestConf) // 初始化RESTful服务器
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
