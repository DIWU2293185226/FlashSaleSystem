// ═════════════════════════════════════════════════════════════════════
// 优惠券服务 — 普通券/秒杀券的增删改查
// 秒杀券需要同时写入 tb_voucher（主表）和 tb_seckill_voucher（扩展表）
// ListByShopID 需要遍历全部分片再聚合结果
// ═════════════════════════════════════════════════════════════════════
package service

import (
	"strconv"

	"github.com/javaup/flashsale-system/internal/dto"
	"github.com/javaup/flashsale-system/internal/model"
	"github.com/javaup/flashsale-system/internal/repository"
	"github.com/javaup/flashsale-system/pkg"
)

// VoucherService 优惠券业务逻辑（普通券 + 秒杀券 CRUD）
type VoucherService struct {
	voucherRepo        *repository.VoucherRepository
	seckillVoucherRepo *repository.SeckillVoucherRepository
	shopRepo           *repository.ShopRepository
}

func NewVoucherService(
	voucherRepo *repository.VoucherRepository,
	seckillVoucherRepo *repository.SeckillVoucherRepository,
	shopRepo *repository.ShopRepository,
) *VoucherService {
	return &VoucherService{
		voucherRepo:        voucherRepo,
		seckillVoucherRepo: seckillVoucherRepo,
		shopRepo:           shopRepo,
	}
}

// AddNormal 新增普通优惠券
func (s *VoucherService) AddNormal(v *model.Voucher) *pkg.Result {
	if err := s.voucherRepo.Create(v); err != nil {
		return pkg.FailWithMsg("新增优惠券失败")
	}
	return pkg.OK()
}

// AddSeckill 新增秒杀优惠券（同时写入 tb_voucher + tb_seckill_voucher）
func (s *VoucherService) AddSeckill(dto *dto.SeckillVoucherDto) *pkg.Result {
	v := &model.Voucher{
		ShopID:      dto.ShopID,
		Title:       dto.Title,
		SubTitle:    dto.SubTitle,
		Rules:       dto.Rules,
		PayValue:    dto.PayValue,
		ActualValue: dto.ActualValue,
		Type:        1, // 秒杀类型
		Status:      1, // 上架
	}
	if err := s.voucherRepo.Create(v); err != nil {
		return pkg.FailWithMsg("新增秒杀优惠券失败")
	}
	sv := &model.SeckillVoucher{
		VoucherID:     v.ID,
		InitStock:     dto.Stock,
		Stock:         dto.Stock,
		AllowedLevels: dto.AllowedLevels,
		MinLevel:      dto.MinLevel,
		BeginTime:     dto.BeginTime,
		EndTime:       dto.EndTime,
	}
	if err := s.seckillVoucherRepo.Create(sv); err != nil {
		return pkg.FailWithMsg("新增秒杀详情失败")
	}
	return pkg.OKWithData(v.ID)
}

// GetByID 查询优惠券，秒杀券附带关联的秒杀信息（库存、时间段）
func (s *VoucherService) GetByID(id int64) *pkg.Result {
	v, err := s.voucherRepo.GetByID(id)
	if err != nil {
		return pkg.FailWithMsg("优惠券不存在")
	}
	if v.Type == 1 {
		sv, err := s.seckillVoucherRepo.GetByVoucherID(id)
		if err == nil {
			v.Stock = sv.Stock
			v.BeginTime = sv.BeginTime
			v.EndTime = sv.EndTime
		}
	}
	return pkg.OKWithData(v)
}

// ListByShopID 查询商铺的优惠券列表
// 由于 shopId 不是分片键，需要遍历全部分片再聚合
func (s *VoucherService) ListByShopID(shopID int64) *pkg.Result {
	var allVouchers []model.Voucher
	for _, dbName := range []string{"hmdp_0", "hmdp_1"} {
		db := s.voucherRepo.GetDB(dbName)
		if db == nil {
			continue
		}
		for i := 0; i < 2; i++ {
			table := "tb_voucher_" + strconv.Itoa(i)
			var vouchers []model.Voucher
			if err := db.Table(table).Where("shop_id = ?", shopID).Find(&vouchers).Error; err == nil {
				allVouchers = append(allVouchers, vouchers...)
			}
		}
	}
	// 遍历填充秒杀信息
	for i := range allVouchers {
		if allVouchers[i].Type == 1 {
			sv, err := s.seckillVoucherRepo.GetByVoucherID(allVouchers[i].ID)
			if err == nil {
				allVouchers[i].Stock = sv.Stock
				allVouchers[i].BeginTime = sv.BeginTime
				allVouchers[i].EndTime = sv.EndTime
			}
		}
	}
	if allVouchers == nil {
		allVouchers = []model.Voucher{}
	}
	return pkg.OKWithData(allVouchers)
}

// UpdateSeckill 更新秒杀优惠券基础信息
func (s *VoucherService) UpdateSeckill(v *model.Voucher) *pkg.Result {
	if err := s.voucherRepo.Update(v); err != nil {
		return pkg.FailWithMsg("更新优惠券失败")
	}
	return pkg.OK()
}

// UpdateSeckillStock 更新秒杀库存（增量更新，diff 为正则增加，负则减少）
func (s *VoucherService) UpdateSeckillStock(voucherID int64, stock int) *pkg.Result {
	if stock < 0 {
		return pkg.FailWithCode(pkg.ErrSeckillStockNotNegative)
	}
	sv, err := s.seckillVoucherRepo.GetByVoucherID(voucherID)
	if err != nil {
		return pkg.FailWithMsg("秒杀优惠券不存在")
	}
	diff := stock - sv.Stock
	if err := s.seckillVoucherRepo.UpdateStockByOffset(voucherID, diff); err != nil {
		return pkg.FailWithMsg("更新库存失败")
	}
	return pkg.OK()
}
