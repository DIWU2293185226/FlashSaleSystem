// ═════════════════════════════════════════════════════════════════════
// 数据库连接管理 — 多数据源初始化
// 支持多个 MySQL 数据源（hmdp_0, hmdp_1），每个连接独立配置连接池
// ═════════════════════════════════════════════════════════════════════
package repository

import (
	"fmt"
	"time"

	"github.com/javaup/flashsale-system/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DatabaseManager 多数据源连接管理器
// 内部用 map 存储数据源名 → *gorm.DB 的映射，上层通过 GetDB(name) 获取
type DatabaseManager struct {
	Sources map[string]*gorm.DB
}

// GetDB 按数据源名称返回对应的 GORM 数据库连接
func (m *DatabaseManager) GetDB(name string) *gorm.DB {
	return m.Sources[name]
}

// InitDatabases 根据配置初始化所有 MySQL 数据源连接
// 遍历 config 中的 DataSource 列表，为每个数据源建立 GORM 连接
// 并为每个连接设置连接池参数（最大打开数、最大空闲数、存活时间等）
func InitDatabases(cfg *config.DatabaseConfig) (*DatabaseManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}

	sources := make(map[string]*gorm.DB, len(cfg.Sources))
	for _, ds := range cfg.Sources {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:                       ds.DSN(),
			DefaultStringSize:         256,
			SkipInitializeWithVersion: false,
		}), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database %s: %w", ds.Name, err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get underlying sql.DB for %s: %w", ds.Name, err)
		}

		// 连接池配置：防止高并发下频繁建连导致 MySQL 连接被打满
		sqlDB.SetMaxOpenConns(ds.MaxOpen)
		sqlDB.SetMaxIdleConns(ds.MaxIdle)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)

		sources[ds.Name] = db
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no database sources configured")
	}

	return &DatabaseManager{Sources: sources}, nil
}

// MustInitDatabases 初始化数据库，失败则 panic
// 用于启动阶段 — 数据库连不上直接退出，不继续运行
func MustInitDatabases(cfg *config.DatabaseConfig) *DatabaseManager {
	dm, err := InitDatabases(cfg)
	if err != nil {
		panic(fmt.Sprintf("database init failed: %v", err))
	}
	return dm
}
