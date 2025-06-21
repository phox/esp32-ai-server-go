package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"ai-server-go/src/api"
	"ai-server-go/src/configs"
	"ai-server-go/src/core"
	"ai-server-go/src/core/auth"
	"ai-server-go/src/core/pool"
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"
	"ai-server-go/src/ota"
	"ai-server-go/src/vision"

	// 导入所有providers以确保init函数被调用
	_ "ai-server-go/src/core/providers/asr/doubao"
	_ "ai-server-go/src/core/providers/asr/gosherpa"
	_ "ai-server-go/src/core/providers/llm/ollama"
	_ "ai-server-go/src/core/providers/llm/openai"
	_ "ai-server-go/src/core/providers/tts/doubao"
	_ "ai-server-go/src/core/providers/tts/edge"
	_ "ai-server-go/src/core/providers/tts/gosherpa"
	_ "ai-server-go/src/core/providers/vlllm/ollama"
	_ "ai-server-go/src/core/providers/vlllm/openai"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

func LoadConfigAndLogger() (*configs.Config, *utils.Logger, error) {
	// 加载配置,默认使用.config.yaml
	config, configPath, err := configs.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	// 初始化日志系统
	logger, err := utils.NewLogger(config)
	if err != nil {
		return nil, nil, err
	}
	logger.Info(fmt.Sprintf("日志系统初始化成功, 配置文件路径: %s", configPath))

	return config, logger, nil
}

func StartWSServer(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context, configService *database.ConfigService) (*core.WebSocketServer, error) {
	// 创建 WebSocket 服务
	wsServer, err := core.NewWebSocketServer(config, logger, configService)
	if err != nil {
		return nil, err
	}

	// 启动 WebSocket 服务
	g.Go(func() error {
		// 监听关闭信号
		go func() {
			<-groupCtx.Done()
			logger.Info("收到关闭信号，开始关闭WebSocket服务...")
			if err := wsServer.Stop(); err != nil {
				logger.Error("WebSocket服务关闭失败", err)
			} else {
				logger.Info("WebSocket服务已优雅关闭")
			}
		}()

		if err := wsServer.Start(groupCtx); err != nil {
			if groupCtx.Err() != nil {
				return nil // 正常关闭
			}
			logger.Error("WebSocket 服务运行失败", err)
			return err
		}
		return nil
	})

	logger.Info("WebSocket 服务已成功启动")
	return wsServer, nil
}

func StartHttpServer(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context, configService *database.ConfigService) (*http.Server, error) {
	// 初始化Gin引擎
	if config.Log.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies([]string{"0.0.0.0"})

	// 初始化数据库连接
	db, err := database.NewDatabase(&config.Database, logger)
	if err != nil {
		logger.Error("数据库连接失败: %v", err)
		return nil, err
	}
	defer db.Close()

	// 初始化服务
	userService := database.NewUserService(db, logger)
	deviceService := database.NewDeviceService(db, logger)
	authMiddleware := auth.NewAuthMiddleware(userService, logger)

	// 从数据库查找 is_default=true 的 provider 作为 defaultModules
	defaultModules, err := configService.GetDefaultProviderModules()
	if err != nil {
		logger.Error("获取默认Provider模块失败: %v", err)
		return nil, err
	}
	// deleteAudio 默认 true
	deleteAudio := true

	// 创建资源池管理器
	poolManager, err := pool.NewPoolManager(config, logger, defaultModules, deleteAudio, configService)
	if err != nil {
		logger.Error("创建资源池管理器失败: %v", err)
		return nil, err
	}
	defer poolManager.Close()

	// 创建用户管理API
	userAPI := api.NewUserAPI(userService, deviceService, configService, authMiddleware, logger, poolManager)
	userAPI.RegisterRoutes(router)

	// API路由全部挂载到/api前缀下
	apiGroup := router.Group("/api")
	// 启动OTA服务
	otaService := ota.NewDefaultOTAService(config.Web.Websocket)
	if err := otaService.Start(groupCtx, router, apiGroup); err != nil {
		logger.Error("OTA 服务启动失败", err)
		return nil, err
	}

	// 启动Vision服务
	visionService, err := vision.NewDefaultVisionService(config, logger, configService)
	if err != nil {
		logger.Error("Vision 服务初始化失败 %v", err)
		return nil, err
	}
	if err := visionService.Start(groupCtx, router, apiGroup); err != nil {
		logger.Error("Vision 服务启动失败", err)
		return nil, err
	}

	// HTTP Server（支持优雅关机）
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Web.Port),
		Handler: router,
	}

	g.Go(func() error {
		logger.Info(fmt.Sprintf("Gin 服务已启动，访问地址: http://0.0.0.0:%d", config.Web.Port))

		// 在单独的 goroutine 中监听关闭信号
		go func() {
			<-groupCtx.Done()
			logger.Info("收到关闭信号，开始关闭HTTP服务...")

			// 创建关闭超时上下文
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP服务关闭失败", err)
			} else {
				logger.Info("HTTP服务已优雅关闭")
			}
		}()

		// ListenAndServe 返回 ErrServerClosed 时表示正常关闭
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP 服务启动失败", err)
			return err
		}
		return nil
	})

	return httpServer, nil
}

func GracefulShutdown(cancel context.CancelFunc, logger *utils.Logger, g *errgroup.Group) {
	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 等待信号
	sig := <-sigChan
	logger.Info(fmt.Sprintf("接收到系统信号: %v，开始优雅关闭服务", sig))

	// 取消上下文，通知所有服务开始关闭
	cancel()

	// 等待所有服务关闭，设置超时保护
	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Error("服务关闭过程中出现错误", err)
			os.Exit(1)
		}
		logger.Info("所有服务已优雅关闭")
	case <-time.After(15 * time.Second):
		logger.Error("服务关闭超时，强制退出")
		os.Exit(1)
	}
}

func startServices(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context) error {
	// 初始化数据库连接
	db, err := database.NewDatabase(&config.Database, logger)
	if err != nil {
		logger.Error("数据库连接失败: %v", err)
		return err
	}
	defer db.Close()

	// 初始化系统配置
	configService := database.NewConfigService(db, logger)
	if err := configService.InitializeDefaultSystemConfigs(); err != nil {
		logger.Error("初始化默认系统配置失败: %v", err)
		return err
	}
	logger.Info("默认系统配置初始化完成")

	// 启动 WebSocket 服务
	if _, err := StartWSServer(config, logger, g, groupCtx, configService); err != nil {
		return fmt.Errorf("启动 WebSocket 服务失败: %w", err)
	}

	// 启动 Http 服务
	if _, err := StartHttpServer(config, logger, g, groupCtx, configService); err != nil {
		return fmt.Errorf("启动 Http 服务失败: %w", err)
	}

	return nil
}

func main() {
	// 加载配置和初始化日志系统
	config, logger, err := LoadConfigAndLogger()
	if err != nil {
		fmt.Println("加载配置或初始化日志系统失败:", err)
		os.Exit(1)
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 用 errgroup 管理两个服务
	g, groupCtx := errgroup.WithContext(ctx)

	// 启动所有服务
	if err := startServices(config, logger, g, groupCtx); err != nil {
		logger.Error("启动服务失败:", err)
		cancel()
		os.Exit(1)
	}

	// 启动优雅关机处理
	GracefulShutdown(cancel, logger, g)

	logger.Info("程序已成功退出")
}
