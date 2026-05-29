// ═════════════════════════════════════════════════════════════════════
// 配置管理模块 — 基于 Viper 的统一配置加载
// 支持：YAML 文件 + 环境变量覆盖（FLASHSALE_ 前缀）
// 默认值兜底，零值自动填充
// ═════════════════════════════════════════════════════════════════════
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 顶层配置结构，聚合所有子模块配置
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Kafka       KafkaConfig       `mapstructure:"kafka"`
	Shard       ShardConfig       `mapstructure:"shard"`
	BloomFilter BloomFilterConfig `mapstructure:"bloom_filter"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Log         LogConfig         `mapstructure:"log"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"` // 监听端口, 默认 8085
	Mode string `mapstructure:"mode"` // debug | release | test
}

// DataSourceConfig 单个 MySQL 数据源连接信息
// DSN() 方法自动拼接连接字符串
type DataSourceConfig struct {
	Name     string `mapstructure:"name"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	MaxOpen  int    `mapstructure:"max_open"`  // 最大连接数
	MaxIdle  int    `mapstructure:"max_idle"`  // 最大空闲数
}

func (d *DataSourceConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Database)
}

// DatabaseConfig 多数据源配置（对应分片后的多个 MySQL 实例/库）
type DatabaseConfig struct {
	Driver  string             `mapstructure:"driver"`  // 驱动类型（仅支持 mysql）
	Sources []DataSourceConfig `mapstructure:"sources"` // 数据源列表，索引对应分片编号
}

// RedisConfig Redis 连接配置
// Addr() 方法组合 host:port
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database int    `mapstructure:"database"` // 选库编号（0-15）
	Password string `mapstructure:"password"`
	PoolSize int    `mapstructure:"pool_size"` // 连接池大小
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// KafkaConfig 消息队列配置
type KafkaConfig struct {
	Brokers         []string `mapstructure:"brokers"`          // Broker 地址列表
	ConsumerGroup   string   `mapstructure:"consumer_group"`   // 消费者组 ID
	AutoOffsetReset string   `mapstructure:"auto_offset_reset"` // 偏移重置策略（latest | earliest）
}

// ShardConfig 分片路由配置
// 控制库和表的数量，所有分片计算基于这两个参数
type ShardConfig struct {
	DbCount    int `mapstructure:"db_count"`    // 数据库个数
	TableCount int `mapstructure:"table_count"` // 每库表个数
}

// BloomFilterConfig 布隆过滤器配置
// 支持为不同业务（商铺、优惠券）设置独立的过滤器参数
type BloomFilterConfig struct {
	Shop    BloomFilterItemConfig `mapstructure:"shop"`
	Voucher BloomFilterItemConfig `mapstructure:"voucher"`
}

// BloomFilterItemConfig 单个布隆过滤器参数
// ExpectedInsertions: 预估插入量，影响位数组大小
// FalseProbability: 期望误判率（越小越精确，但占用更多空间）
type BloomFilterItemConfig struct {
	Name               string  `mapstructure:"name"`
	ExpectedInsertions uint    `mapstructure:"expected_insertions"`
	FalseProbability   float64 `mapstructure:"false_probability"`
}

// JWTConfig 用户认证令牌配置
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`           // 签名密钥
	ExpirationHours int    `mapstructure:"expiration_hours"` // 过期时间（小时）
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug | info | warn | error
	Output string `mapstructure:"output"` // 日志输出路径（空=stdout）
}

// Load 从 YAML 文件加载配置，支持环境变量覆盖
// 环境变量前缀 FLASHSALE，层级用下划线分隔
// 例: FLASHSALE_SERVER_PORT=9090 覆盖 server.port
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix("FLASHSALE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 关键字段零值兜底，防止空配置导致服务异常
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8085
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Shard.DbCount == 0 {
		cfg.Shard.DbCount = 2
	}
	if cfg.Shard.TableCount == 0 {
		cfg.Shard.TableCount = 2
	}

	return &cfg, nil
}

// MustLoad 加载配置，失败时直接 panic（适用于启动阶段快速失败）
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("config load failed: %v", err))
	}
	return cfg
}
