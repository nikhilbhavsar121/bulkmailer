package worker

import (
	"sync"

	"bulkmailer/internal/limiter"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const (
	TypeEmailSend = "email:send"
	QueueDefault  = "default"
	GlobalRateKey = "GLOBAL"
)

type EmailPayload struct {
	To      string
	Subject string
	Body    string
}

type Worker struct {
	AsynqClient *asynq.Client
	RedisClient *redis.Client
	Inspector   *asynq.Inspector
	RateLimiter *limiter.SlidingRateLimiter
	CSVPath     string

	serverMu      sync.Mutex
	asynqSrv      *asynq.Server
	sentCount     int64
	failedCount   int64
	enqueuedCount int64
	statsMu       sync.Mutex
}

func NewWorker(c *asynq.Client, r *redis.Client, i *asynq.Inspector, l *limiter.SlidingRateLimiter, csv string) *Worker {
	return &Worker{
		AsynqClient: c,
		RedisClient: r,
		Inspector:   i,
		RateLimiter: l,
		CSVPath:     csv,
	}
}
