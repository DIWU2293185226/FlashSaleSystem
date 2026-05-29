package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/javaup/flashsale-system/internal/cache"
	"github.com/javaup/flashsale-system/internal/mq"
	"github.com/redis/go-redis/v9"
)

// Ensure context import is used
var _ = context.Background

// DelayQueueService provides a Redis-based delayed queue for voucher reminders.
type DelayQueueService struct {
	redis *cache.RedisCache
	producer *mq.Producer
}

// NewDelayQueueService creates a new delay queue service.
func NewDelayQueueService(redis *cache.RedisCache, producer *mq.Producer) *DelayQueueService {
	return &DelayQueueService{redis: redis, producer: producer}
}

const delayQueueKey = "delay:queue:voucher"

// AddReminder adds a delayed reminder. The message will be available after delaySeconds.
func (s *DelayQueueService) AddReminder(voucherID int64, userID int64, delaySeconds int) error {
	ctx := context.Background()
	score := float64(time.Now().UnixMilli() + int64(delaySeconds)*1000)

	member := strconv.FormatInt(voucherID, 10) + ":" + strconv.FormatInt(userID, 10)
	return s.redis.Client.ZAdd(ctx, delayQueueKey, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// PollDueReminders fetches and processes all due reminders.
// This should be called periodically (e.g., every second via a cron-like loop).
func (s *DelayQueueService) PollDueReminders() ([]mq.DelayReminderMessage, error) {
	ctx := context.Background()
	now := float64(time.Now().UnixMilli())

	// Fetch due items
	members, err := s.redis.Client.ZRangeByScore(ctx, delayQueueKey, &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatFloat(now, 'f', 0, 64),
		Offset: 0,
		Count:  100,
	}).Result()
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, nil
	}

	// Remove fetched items
	_, _ = s.redis.Client.ZRem(ctx, delayQueueKey, convertToInterface(members)...).Result()

	var reminders []mq.DelayReminderMessage
	for _, member := range members {
		voucherID, userID := parseMember(member)
		if voucherID > 0 {
			reminders = append(reminders, mq.DelayReminderMessage{
				VoucherID: voucherID,
				UserID:    userID,
			})
		}
	}

	// Send to Kafka for processing
	for _, r := range reminders {
		_ = s.producer.Send(ctx, "delay_voucher_reminder",
			fmt.Sprintf("%d_%d", r.VoucherID, r.UserID), r)
	}

	return reminders, nil
}

// RemoveReminder removes a pending reminder.
func (s *DelayQueueService) RemoveReminder(voucherID, userID int64) error {
	ctx := context.Background()
	member := strconv.FormatInt(voucherID, 10) + ":" + strconv.FormatInt(userID, 10)
	return s.redis.Client.ZRem(ctx, delayQueueKey, member).Err()
}

func parseMember(member string) (voucherID, userID int64) {
	parts := split2(member, ":")
	if len(parts) == 2 {
		voucherID, _ = strconv.ParseInt(parts[0], 10, 64)
		userID, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return
}

func split2(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func convertToInterface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
