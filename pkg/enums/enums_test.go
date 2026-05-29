package enums

import (
	"testing"
)

func TestOrderStatus(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		code     int
		str      string
	}{
		{OrderStatusNormal, 1, "正常"},
		{OrderStatusCancel, 2, "取消"},
		{OrderStatus(999), 999, "未知"},
	}

	for _, tt := range tests {
		if got := tt.status.Code(); got != tt.code {
			t.Errorf("OrderStatus(%d).Code() = %d, want %d", tt.status, got, tt.code)
		}
		if got := tt.status.String(); got != tt.str {
			t.Errorf("OrderStatus(%d).String() = %s, want %s", tt.status, got, tt.str)
		}
	}
}

func TestVoucherStatus(t *testing.T) {
	tests := []struct {
		status VoucherStatus
		code   int
		str    string
	}{
		{VoucherStatusAvailable, 1, "上架"},
		{VoucherStatusUnavailable, 2, "下架"},
		{VoucherStatusExpired, 3, "过期"},
	}

	for _, tt := range tests {
		if got := tt.status.Code(); got != tt.code {
			t.Errorf("VoucherStatus(%d).Code() = %d, want %d", tt.status, got, tt.code)
		}
		if got := tt.status.String(); got != tt.str {
			t.Errorf("VoucherStatus(%d).String() = %s, want %s", tt.status, got, tt.str)
		}
	}
}

func TestSubscribeStatus(t *testing.T) {
	tests := []struct {
		status SubscribeStatus
		code   int
		str    string
	}{
		{SubscribeStatusUnsubscribed, 0, "未订阅"},
		{SubscribeStatusSubscribed, 1, "已订阅"},
		{SubscribeStatusSuccess, 2, "已成功"},
	}

	for _, tt := range tests {
		if got := tt.status.Code(); got != tt.code {
			t.Errorf("SubscribeStatus(%d).Code() = %d, want %d", tt.status, got, tt.code)
		}
		if got := tt.status.String(); got != tt.str {
			t.Errorf("SubscribeStatus(%d).String() = %s, want %s", tt.status, got, tt.str)
		}
	}
}
