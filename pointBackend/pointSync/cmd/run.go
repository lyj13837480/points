package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"pointSync/internal/config"
	"pointSync/internal/logger/xzap"
	pService "pointSync/internal/service"
	"pointSync/internal/service/point"
	"pointSync/internal/stores/gdb"
	"pointSync/internal/stores/xmq"

	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/service"
	"go.uber.org/zap"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "sync easy swap order info.",
	Long:  "sync easy swap order info.",
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		// rpc退出信号通知chan
		onSyncExit := make(chan error, 1)

		go func() {
			defer wg.Done()

			c, err := config.UnmarshalCmdConfig() // 读取和解析配置文件
			if err != nil {
				xzap.WithContext(ctx).Error("Failed to unmarshal config", zap.Error(err))
				onSyncExit <- err
				return
			}

			fmt.Printf("config %v\n", c)
			_, err = xzap.SetUp(*c.AppLog) // 初始化日志模块
			if err != nil {
				xzap.WithContext(ctx).Error("Failed to set up logger", zap.Error(err))
				onSyncExit <- err
				return
			}
			gdb.MustNewDB(&c.Mysql) // 初始化数据库连接

			fmt.Println("initConfig done")
			// 初始化服务上下文
			serviceCtx := pService.New(rootCmd.Context(), c)
			// 启动服务上下文
			serviceCtx.Start()

			ctx := context.Background()
			serviceGroup := service.NewServiceGroup()
			defer serviceGroup.Stop()

			producer := xmq.New(c)

			pointService := point.New(producer.KqPusherClient, ctx, c.KqPusherConf, gdb.DB)
			// 启动推送积分任务
			pointService.Start()

			for _, mq := range xmq.Consumers(c, ctx) {
				serviceGroup.Add(mq)
			}

			serviceGroup.Start()

			xzap.WithContext(ctx).Info("sync server start", zap.Any("config", c))

		}()

		// 信号通知chan
		onSignal := make(chan os.Signal)
		// 优雅退出
		signal.Notify(onSignal, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-onSignal:
			switch sig {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				cancel()
				xzap.WithContext(ctx).Info("Exit by signal", zap.String("signal", sig.String()))
			}
		case err := <-onSyncExit:
			cancel()
			xzap.WithContext(ctx).Error("Exit by error", zap.Error(err))
		}
		wg.Wait()
	},
}

func init() {
	// 将api初始化命令添加到主命令中
	rootCmd.AddCommand(DaemonCmd)
}
