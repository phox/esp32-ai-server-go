package pool_test

import (
	"ai-server-go/src/configs"
	"ai-server-go/src/core/pool"
	"ai-server-go/src/core/utils"
	"ai-server-go/src/database"
	"context"
	"testing"
)

func TestConnectivityCheck(t *testing.T) {
	config, _, err := configs.LoadConfig()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	logger, err := utils.NewLogger(config)
	if err != nil {
		t.Fatalf("创建日志记录器失败: %v", err)
	}
	configService := database.NewConfigService(nil, logger)
	connConfig := pool.DefaultConnectivityConfig()
	healthChecker := pool.NewHealthChecker(config, configService, connConfig, logger)
	ctx := context.Background()
	if err := healthChecker.CheckAllProviders(ctx, pool.BasicCheck, nil); err != nil {
		t.Errorf("基础连通性检查失败: %v", err)
	}
	if err := healthChecker.CheckAllProviders(ctx, pool.FunctionalCheck, nil); err != nil {
		t.Errorf("功能性检查失败: %v", err)
	}
}
