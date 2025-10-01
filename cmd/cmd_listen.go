package cmd

import (
	"context"
	"fmt"
	"github.com/chihqiang/dbxgo/config"
	"github.com/chihqiang/dbxgo/output"
	"github.com/chihqiang/dbxgo/source"
	"github.com/chihqiang/dbxgo/store"
	"github.com/urfave/cli/v3"
	"log/slog"
	"os"
	"runtime"
	"sync"
)

func ListenCommand() *cli.Command {
	return &cli.Command{
		UseShortOptionHandling: true,
		Name:                   "listen",
		Usage:                  "Listen to CDC events without sending them to any output",
		Flags:                  []cli.Flag{},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// 从context中获取配置
			cfg, ok := ctx.Value(ContextValueConfig).(*config.Config)
			if !ok {
				return fmt.Errorf("config not found in context")
			}
			return Listen(ctx, cfg)
		},
	}
}

// Listen 启动整个 CDC 监听流程，包括数据源、存储和输出处理
func Listen(ctx context.Context, config *config.Config) error {
	// 创建 Store 实例
	iStore, err := store.NewStore(config.Store)
	if err != nil {
		return err // 返回创建 Store 错误
	}
	// 创建数据源实例
	iSource, err := source.NewSource(config.Source)
	if err != nil {
		return err // 返回创建 Source 错误
	}
	// 将 Store 注入到数据源
	iSource.WithStore(iStore)
	// 创建输出实例
	iOutput, err := output.NewOutput(config.Output)
	if err != nil {
		return err // 返回创建 Output 错误
	}

	// 创建可取消上下文，用于控制 goroutine 生命周期
	ctx, cancel := context.WithCancel(ctx)
	// 定义关闭回调函数，统一关闭资源
	closeCallback := func() {
		slog.Info("closing source and output")
		if err := iSource.Close(); err != nil {
			slog.Error("failed to close source", "error", err)
		}
		if err := iOutput.Close(); err != nil {
			slog.Error("failed to close output", "error", err)
		}
		cancel() // 取消上下文
	}

	defer closeCallback() // 函数结束时自动调用关闭回调

	// 创建用于监控数据源运行错误的通道
	sourceErrChan := make(chan error, 1)

	// 启动 goroutine 运行数据源
	go func() {
		slog.Info("starting source run goroutine")
		if err := iSource.Run(ctx); err != nil {
			// 数据源运行失败时记录错误并发送到通道
			slog.Error("source run failed", "error", err)
			sourceErrChan <- err
		}
		close(sourceErrChan) // 数据源结束后关闭通道
	}()

	// 定义 worker 数量，使用CPU核心数
	workerCount := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// 启动 worker goroutine 处理数据源事件
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			defer wg.Done() // goroutine 结束时通知 WaitGroup
			slog.Info("worker started", "workerID", id)
			for {
				select {
				case event, ok := <-iSource.GetChanEventData():
					// 读取事件通道
					if !ok {
						// 通道关闭，worker 退出
						slog.Info("event channel closed", "workerID", id)
						return
					}
					slog.Info("CDC Event", slog.Any("event", event))
					// 发送事件到输出
					if err := output.SendWithRetry(ctx, iOutput, event, 3); err != nil {
						slog.Error("failed to send event", "workerID", id, "error", err)
					}
				case <-ctx.Done():
					// 上下文取消，worker 退出
					slog.Info("context canceled, worker exiting", "workerID", id)
					return
				}
			}
		}(i)
	}
	// 等待数据源错误或 worker 完成
	err, ok := <-sourceErrChan
	if ok && err != nil {
		// 数据源运行错误时，关闭所有资源并等待 worker 退出
		closeCallback()
		wg.Wait()
		slog.Error("exiting program due to source run error", "error", err)
		os.Exit(1)
	}
	// 数据源正常结束，等待 worker 完成
	wg.Wait()
	return nil
}
