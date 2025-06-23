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
	fmt.Println(configPath)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(config)

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

func StartHttpServer(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context, configService *database.ConfigService, db *database.Database) (*http.Server, error) {
	// 初始化Gin引擎
	if config.Log.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies([]string{"0.0.0.0"})

	// 添加根路径路由用于测试
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI Server Go is running",
			"version": "1.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 执行数据库自动迁移
	if err := db.AutoMigrate(); err != nil {
		logger.Error("数据库迁移失败: %v", err)
		return nil, err
	}

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

	// API路由全部挂载到/api前缀下
	apiGroup := router.Group("/api")

	// 创建用户管理API
	userAPI := api.NewUserAPI(userService, deviceService, configService, authMiddleware, logger, poolManager)
	userAPI.RegisterRoutes(apiGroup)

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

	// 输出所有已注册路由（在所有服务注册完成后）
	logger.Info("=== 所有已注册的路由 ===")
	for _, ri := range router.Routes() {
		logger.Info(fmt.Sprintf("已注册路由: %s %s -> %s", ri.Method, ri.Path, ri.Handler))
	}
	logger.Info("=== 路由注册完成 ===")

	// 检查是否存在管理员账号，不存在则创建
	logger.Info("检查是否存在管理员账号，不存在则创建")
	adminUser, err := userService.GetUserByUsername("admin")

	if err != nil || adminUser == nil {
		admin := &database.User{
			Username: "admin",
			Nickname: "管理员",
			Status:   "active",
			Role:     "admin",
		}
		err = userService.CreateUser(admin, "admin123")
		if err != nil {
			logger.Error("创建默认管理员账号失败: %v", err)
		} else {
			logger.Info("已自动创建默认管理员账号: admin/admin123")
		}
	}

	// HTTP Server（支持优雅关机）
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Web.Port),
		Handler: router,
	}

	logger.Info(fmt.Sprintf("准备启动HTTP服务，监听端口: %d", config.Web.Port))

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

	// 等待所有服务关闭完成
	if err := g.Wait(); err != nil {
		logger.Error("服务关闭过程中发生错误", err)
	} else {
		logger.Info("所有服务已优雅关闭")
	}
}

func main() {
	// 加载配置和日志
	config, logger,
		err := LoadConfigAndLogger()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("AI服务器启动中...")

	// 初始化数据库连接
	db, err := database.NewDatabase(&config.Database, logger)
	if err != nil {
		logger.Error("数据库连接失败: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// 创建错误组和上下文
	g, groupCtx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(groupCtx)

	// 初始化配置服务
	configService := database.NewConfigService(db, logger)
	// 初始化默认系统配置
	if err := configService.InitializeDefaultSystemConfigs(); err != nil {
		logger.Error("初始化默认系统配置失败: %v", err)
		os.Exit(1)
	}
	// 初始化默认provider配置
	if err := configService.InitializeDefaultProviderConfigs(); err != nil {
		logger.Error("初始化默认provider配置失败: %v", err)
		os.Exit(1)
	}

	// 启动WebSocket服务
	_, err = StartWSServer(config, logger, g, ctx, configService)
	if err != nil {
		logger.Error("启动WebSocket服务失败", err)
		os.Exit(1)
	}

	// 启动HTTP服务（内部完成所有服务注册和初始化）
	_, err = StartHttpServer(config, logger, g, ctx, configService, db)
	if err != nil {
		logger.Error("启动服务失败", err)
		os.Exit(1)
	}

	// 优雅关闭
	GracefulShutdown(cancel, logger, g)
}
