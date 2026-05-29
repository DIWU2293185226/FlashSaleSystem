// ═════════════════════════════════════════════════════════════════════
// 应用组装 — 依赖注入 + Gin 引擎 + 路由注册
// NewApp 负责创建所有组件并将它们 wired 在一起
// SetupRouter 注册全部 49+ 条路由，按模块分组
// 启动流程：main → NewApp → SetupRouter → Run
// 关闭流程：signal → Shutdown → 释放资源
// ═════════════════════════════════════════════════════════════════════
package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/javaup/flashsale-system/internal/bloom"
	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/config"
	"github.com/javaup/flashsale-system/internal/handler"
	"github.com/javaup/flashsale-system/internal/idgen"
	"github.com/javaup/flashsale-system/internal/jwt"
	"github.com/javaup/flashsale-system/internal/locker"
	"github.com/javaup/flashsale-system/internal/middleware"
	"github.com/javaup/flashsale-system/internal/mq"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/internal/service"
	"github.com/javaup/flashsale-system/internal/sharding"
	"github.com/rs/zerolog"
)

// App 应用容器，持有所有组件的引用
// 包括配置、数据库、缓存、消息队列、业务服务和 HTTP 处理器
// 同时也持有需要生命周期管理的内部组件（producer、consumer、snowflake）
type App struct {
	Config          *config.Config
	DB              *repository.DatabaseManager
	Redis           *cache.RedisCache
	Log             zerolog.Logger
	Engine          *gin.Engine
	HealthHandler   *handler.HealthHandler
	UserHandler     *handler.UserHandler
	ShopHandler     *handler.ShopHandler
	ShopTypeHandler *handler.ShopTypeHandler
	VoucherHandler  *handler.VoucherHandler
	BlogHandler     *handler.BlogHandler
	FollowHandler   *handler.FollowHandler
	UploadHandler   *handler.UploadHandler
	SeckillHandler  *handler.SeckillHandler
	SubscribeHandler *handler.SubscribeHandler
	ReconcileHandler *handler.ReconciliationHandler

	// 内部组件 — 持有引用以便优雅关闭时释放资源
	luaManager  *cache.LuaManager
	producer    *mq.Producer
	consumer    *mq.Consumer
	idGen       *idgen.Snowflake
}

// NewApp 组装所有依赖，完成依赖注入
// 初始化顺序：DB → Redis → 分片路由 → JWT → 仓库层 → 缓存层 → Lua脚本 → Snowflake
//          → 分布式锁 → 布隆过滤器 → Kafka生产者 → 业务服务 → HTTP处理器 → Kafka消费者
// 所有组件的生命周期由 App 统一管理
func NewApp(cfg *config.Config, log zerolog.Logger) *App {
	// 初始化数据库连接（多数据源）
	db := repository.MustInitDatabases(&cfg.Database)

	// 初始化 Redis 连接
	rdb := cache.MustInitRedis(&cfg.Redis)

	// 初始化分片路由器
	router := sharding.NewRouter(&cfg.Shard)

	// 初始化 JWT 管理器
	jwtManager := jwt.NewManager(&cfg.JWT)

	// 初始化所有 Repository（数据访问层）
	userRepo := repository.NewUserRepository(db, router)
	userInfoRepo := repository.NewUserInfoRepository(db, router)
	userPhoneRepo := repository.NewUserPhoneRepository(db, router)
	voucherRepo := repository.NewVoucherRepository(db, router)
	seckillVoucherRepo := repository.NewSeckillVoucherRepository(db, router)
	orderRepo := repository.NewVoucherOrderRepository(db, router)
	shopRepo := repository.NewShopRepository(db, router)
	shopTypeRepo := repository.NewShopTypeRepository(db)
	blogRepo := repository.NewBlogRepository(db)
	blogCommentsRepo := repository.NewBlogCommentsRepository(db)
	followRepo := repository.NewFollowRepository(db)
	reconcileLogRepo := repository.NewVoucherReconcileLogRepository(db, router)
	subscribeRepo := repository.NewSubscribeRepository(db)

	_ = blogCommentsRepo
	_ = voucherRepo

	// 初始化多级缓存（LocalCache → Redis）
	multiCache := cache.NewMultiLevelCache(0, rdb)

	// 初始化 Lua 脚本管理器，加载所有 Lua 脚本到 Redis
	// Lua 脚本是秒杀原子操作的核心，加载失败只会警告不影响启动
	luaDir := "resource/lua"
	if _, err := os.Stat(luaDir); os.IsNotExist(err) {
		luaDir = "../resource/lua"
	}
	luaManager := cache.NewLuaManager(rdb, luaDir)
	if err := luaManager.LoadAll(); err != nil {
		log.Warn().Err(err).Msg("failed to load Lua scripts, seckill may not work")
	}

	// 初始化 Snowflake ID 生成器
	// 通过 Redis Lua 脚本在 0~31 之间自动分配 WorkerID，避免多实例冲突
	var idGen *idgen.Snowflake
	workerID := int64(0)
	datacenterID := int64(0)

	workerKey := "snowflake:worker:" + cfg.Server.Mode
	result, err := luaManager.Eval("workAndDataCenterId",
		[]string{workerKey},
		0, 0,
	)
	if err == nil {
		if arr, ok := result.([]interface{}); ok && len(arr) >= 3 {
			if success, _ := arr[0].(int64); success == 1 {
				workerID, _ = arr[1].(int64)
				datacenterID, _ = arr[2].(int64)
			}
		}
	}
	idGen = idgen.NewSnowflake(workerID, datacenterID)
	log.Info().Int64("workerID", workerID).Int64("datacenterID", datacenterID).Msg("snowflake initialized")

	// 初始化 Redis 分布式锁
	lockerSvc := locker.New(rdb)

	// 初始化布隆过滤器（用于秒杀 voucher 存在性预检）
	bloomFilter := bloom.NewFilter(
		cfg.BloomFilter.Voucher.Name,
		cfg.BloomFilter.Voucher.ExpectedInsertions,
		cfg.BloomFilter.Voucher.FalseProbability,
		rdb,
	)
	_ = bloomFilter

	// 初始化 Kafka 生产者（异步秒杀订单落库）
	producer := mq.NewProducer(cfg.Kafka.Brokers, log)

	// 初始化所有 Service（业务逻辑层）
	userSvc := service.NewUserService(userRepo, userInfoRepo, userPhoneRepo, rdb, jwtManager)
	shopSvc := service.NewShopService(shopRepo, shopTypeRepo, multiCache)
	shopTypeSvc := service.NewShopTypeService(shopTypeRepo)
	voucherSvc := service.NewVoucherService(voucherRepo, seckillVoucherRepo, shopRepo)
	blogSvc := service.NewBlogService(blogRepo, userRepo, followRepo, rdb)
	followSvc := service.NewFollowService(followRepo, userRepo)
	uploadSvc := service.NewUploadService("./uploads")
	seckillSvc := service.NewSeckillService(
		idGen, seckillVoucherRepo, voucherRepo, orderRepo,
		rdb, luaManager, lockerSvc, bloomFilter, producer,
	)
	subscribeSvc := service.NewSubscribeService(subscribeRepo, seckillVoucherRepo, orderRepo, rdb)
	reconcileSvc := service.NewReconciliationService(seckillVoucherRepo, reconcileLogRepo, rdb)
	delayQueueSvc := service.NewDelayQueueService(rdb, producer)
	_ = delayQueueSvc // 延迟队列后台轮询由单独的 goroutine 驱动

	// 初始化所有 Handler（HTTP 处理器）
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userSvc)
	shopHandler := handler.NewShopHandler(shopSvc)
	shopTypeHandler := handler.NewShopTypeHandler(shopTypeSvc)
	voucherHandler := handler.NewVoucherHandler(voucherSvc)
	blogHandler := handler.NewBlogHandler(blogSvc)
	followHandler := handler.NewFollowHandler(followSvc)
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	seckillHandler := handler.NewSeckillHandler(seckillSvc)
	subscribeHandler := handler.NewSubscribeHandler(subscribeSvc)
	reconcileHandler := handler.NewReconciliationHandler(reconcileSvc)

	// 初始化 Kafka 消费者（启动后台 goroutine 消费秒杀订单消息）
	consumer := mq.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, log)
	consumer.ConsumeSeckillOrders(context.Background(), seckillSvc.HandleSeckillOrder)

	return &App{
		Config:           cfg,
		DB:               db,
		Redis:            rdb,
		Log:              log,
		idGen:            idGen,
		luaManager:       luaManager,
		producer:         producer,
		consumer:         consumer,
		HealthHandler:    healthHandler,
		UserHandler:      userHandler,
		ShopHandler:      shopHandler,
		ShopTypeHandler:  shopTypeHandler,
		VoucherHandler:   voucherHandler,
		BlogHandler:      blogHandler,
		FollowHandler:    followHandler,
		UploadHandler:    uploadHandler,
		SeckillHandler:   seckillHandler,
		SubscribeHandler: subscribeHandler,
		ReconcileHandler: reconcileHandler,
	}
}

// SetupRouter 配置 Gin 路由引擎，注册所有中间件和路由
// 中间件顺序：CORS → 请求日志 → Panic 恢复 → JWT 认证（按需）
// 模块路由：health / user / shop / voucher / blog / follow / upload / seckill / subscribe / reconcile
func (app *App) SetupRouter() *gin.Engine {
	ginMode := app.Config.Server.Mode
	switch ginMode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// 全局中间件：CORS 跨域 + 请求日志 + Panic 恢复
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(app.Log))
	r.Use(middleware.Recovery(app.Log))

	// 健康检查（不经过认证中间件）
	r.GET("/ping", app.HealthHandler.Ping)

	// JWT 认证中间件（懒加载，按路由组按需使用）
	jwtManager := jwt.NewManager(&app.Config.JWT)
	authRequired := middleware.AuthRequired(jwtManager)

	// ── 用户模块 ──────────────────────────────────────────────
	user := r.Group("/user")
	{
		user.POST("/code", app.UserHandler.SendCode)
		user.POST("/login", app.UserHandler.Login)
		user.POST("/logout", app.UserHandler.Logout)
		user.GET("/me", authRequired, app.UserHandler.GetMe)
		user.GET("/:id", app.UserHandler.GetByID)
		user.GET("/info/:id", app.UserHandler.GetInfoByID)
		user.POST("/sign", authRequired, app.UserHandler.Sign)
		user.GET("/sign/count", authRequired, app.UserHandler.SignCount)
	}

	// ── 商铺模块 ──────────────────────────────────────────────
	shop := r.Group("/shop")
	{
		shop.GET("/:id", app.ShopHandler.GetByID)
		shop.POST("", app.ShopHandler.Create)
		shop.PUT("", app.ShopHandler.Update)
		shop.GET("/of/type", app.ShopHandler.ListByType)
		shop.GET("/of/name", app.ShopHandler.ListByName)
		shop.GET("/of/nearby", app.ShopHandler.ListNearby)
	}

	// ── 商铺类型 ──────────────────────────────────────────────
	r.GET("/shop-type/list", app.ShopTypeHandler.ListAll)

	// ── 优惠券模块 ────────────────────────────────────────────
	voucher := r.Group("/voucher")
	{
		voucher.POST("", app.VoucherHandler.AddNormal)
		voucher.POST("/seckill", app.VoucherHandler.AddSeckill)
		voucher.POST("/get", app.VoucherHandler.GetByID)
		voucher.GET("/list/:shopId", app.VoucherHandler.ListByShopID)
		voucher.POST("/update/seckill", app.VoucherHandler.UpdateSeckill)
		voucher.POST("/update/seckill/stock", app.VoucherHandler.UpdateSeckillStock)
	}

	// ── 博客模块 ──────────────────────────────────────────────
	blog := r.Group("/blog")
	{
		blog.POST("", authRequired, app.BlogHandler.Create)
		blog.PUT("/like/:id", authRequired, app.BlogHandler.Like)
		blog.GET("/hot", app.BlogHandler.ListHot)
		blog.GET("/:id", app.BlogHandler.GetByID)
		blog.GET("/likes/:id", app.BlogHandler.ListLikes)
		blog.GET("/of/user", app.BlogHandler.ListByUserID)
		blog.GET("/of/me", authRequired, app.BlogHandler.ListMyBlogs)
		blog.GET("/of/follow", authRequired, app.BlogHandler.ListFollowBlog)
	}

	// ── 关注模块 ──────────────────────────────────────────────
	follow := r.Group("/follow")
	{
		follow.PUT("/:id/:isFollow", authRequired, app.FollowHandler.Follow)
		follow.GET("/or/not/:id", authRequired, app.FollowHandler.IsFollowed)
		follow.GET("/common/:id", authRequired, app.FollowHandler.GetCommon)
	}

	// ── 文件上传 ──────────────────────────────────────────────
	upload := r.Group("/upload")
	{
		upload.POST("/blog", authRequired, app.UploadHandler.UploadBlog)
		upload.DELETE("/blog/delete", authRequired, app.UploadHandler.DeleteBlog)
	}

	// ── 秒杀订单（Phase 3）────────────────────────────────────
	voucherOrder := r.Group("/voucher-order")
	{
		voucherOrder.POST("/seckill", authRequired, app.SeckillHandler.Seckill)
		voucherOrder.GET("/:orderId", app.SeckillHandler.GetOrder)
		voucherOrder.PUT("/cancel/:orderId", authRequired, app.SeckillHandler.CancelOrder)
		voucherOrder.POST("/load-stock", app.SeckillHandler.LoadStock)
		voucherOrder.GET("/stock/:voucherId", app.SeckillHandler.GetStock)
		voucherOrder.GET("/voucher/:voucherId", app.SeckillHandler.GetSeckillVoucherFull)
		voucherOrder.POST("/token", authRequired, app.SeckillHandler.GenerateToken)
	}

	// ── 订阅候补（Phase 4）────────────────────────────────────
	subscribe := r.Group("/subscribe")
	{
		subscribe.POST("", authRequired, app.SubscribeHandler.Subscribe)
		subscribe.DELETE("/:voucherId", authRequired, app.SubscribeHandler.Unsubscribe)
		subscribe.GET("/status/:voucherId", authRequired, app.SubscribeHandler.GetStatus)
		subscribe.GET("/history", authRequired, app.SubscribeHandler.GetHistory)
	}

	// ── 库存对账（Phase 4）────────────────────────────────────
	reconcile := r.Group("/reconcile")
	{
		reconcile.GET("/check/:voucherId", app.ReconcileHandler.CheckStock)
		reconcile.POST("/fix/:voucherId", app.ReconcileHandler.FixStock)
	}

	return r
}

// Run 启动 HTTP 服务器，监听配置的端口
func (app *App) Run() error {
	app.Engine = app.SetupRouter()
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	app.Log.Info().Str("addr", addr).Msg("starting server")
	return app.Engine.Run(addr)
}

// Shutdown 优雅关闭，释放所有资源
// 关闭顺序：Kafka 生产者 → Kafka 消费者 → Redis 连接
// 确保在进程退出前完成所有正在处理的消息
func (app *App) Shutdown() {
	if app.producer != nil {
		app.producer.Close()
	}
	if app.consumer != nil {
		app.consumer.Close()
	}
	if app.Redis != nil {
		app.Redis.Close()
	}
	app.Log.Info().Msg("server shutdown complete")
}
