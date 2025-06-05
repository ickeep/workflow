package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/workflow-engine/workflow-engine/internal/temporal"
	"github.com/workflow-engine/workflow-engine/pkg/config"
)

func main() {
	log.Println("启动Temporal Worker...")

	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建Temporal客户端
	temporalClient, err := temporal.NewClient(cfg.Temporal)
	if err != nil {
		log.Fatalf("创建Temporal客户端失败: %v", err)
	}
	defer temporalClient.Close()

	log.Printf("Temporal客户端连接成功: %s", cfg.Temporal.HostPort)
	log.Printf("任务队列: %s", cfg.Temporal.TaskQueue)

	// 启动Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 在单独的goroutine中启动Worker
	go func() {
		if err := temporalClient.StartWorker(ctx); err != nil {
			log.Fatalf("启动Temporal Worker失败: %v", err)
		}
	}()

	log.Println("Temporal Worker 启动成功，等待任务...")

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-sigCh
	log.Println("收到中断信号，正在关闭Worker...")

	// 取消上下文，优雅关闭Worker
	cancel()

	log.Println("Temporal Worker 已关闭")
}
