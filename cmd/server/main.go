// Package main HTTP服务器启动程序
// 工作流引擎HTTP API服务器入口
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/server"
)

func main() {
	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	defer logger.Sync()

	logger.Info("工作流引擎服务启动中...")

	// 创建HTTP路由器
	router := server.NewRouter(logger)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("HTTP服务器启动", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭HTTP服务器...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP服务器关闭失败", zap.Error(err))
	} else {
		logger.Info("HTTP服务器已关闭")
	}
}
