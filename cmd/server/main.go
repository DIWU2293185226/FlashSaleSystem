// ═════════════════════════════════════════════════════════════════════
// 服务启动入口 — 配置加载 → 日志初始化 → 应用组装 → 优雅关闭
// 支持 SIGINT/SIGTERM 信号捕获，关闭前完成正在处理的请求和消息
// ═════════════════════════════════════════════════════════════════════
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/javaup/flashsale-system/internal"
	"github.com/javaup/flashsale-system/internal/config"
	"github.com/rs/zerolog"
)

func main() {
	// 初始化日志（控制台输出，带时间戳和服务名）
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Str("service", "flashsale-system").
		Logger()

	// 加载 YAML 配置文件（数据库/Redis/Kafka/分片等全部配置）
	cfg, err := config.Load("resource/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// 根据配置设置日志级别
	switch cfg.Log.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// 创建应用实例（内部完成所有依赖的初始化）
	app := internal.NewApp(cfg, log)

	// 信号监听，用于优雅关闭
	// 收到 SIGINT（Ctrl+C）或 SIGTERM（kill）时触发关闭流程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 后台 goroutine 启动 HTTP 服务
	go func() {
		if err := app.Run(); err != nil {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	log.Info().Msg("server started successfully")

	// 阻塞等待关闭信号
	<-quit
	log.Info().Msg("shutting down server...")

	// 执行优雅关闭（关闭 Kafka 生产者/消费者、Redis 连接）
	app.Shutdown()
}
