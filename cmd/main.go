package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bulkmailer/internal/api"
	"bulkmailer/internal/limiter"
	"bulkmailer/internal/utils"
	"bulkmailer/internal/worker"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

var (
	RedisAddr = utils.MustGetenv("REDIS_ADDR", "127.0.0.1:6379")
	CSVPath   = utils.MustGetenv("CSV_PATH", "recipients.csv")
)

func main() {
	redisClient := redis.NewClient(&redis.Options{Addr: RedisAddr})
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: RedisAddr})
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: RedisAddr})
	limiter := limiter.NewSlidingRateLimiter(1 * time.Second)

	app := worker.NewWorker(asynqClient, redisClient, inspector, limiter, CSVPath)

	apiHandler := api.NewAPI(app)
	http.HandleFunc("/start", apiHandler.HandleStart)
	http.HandleFunc("/stop", apiHandler.HandleStop)
	http.HandleFunc("/pause", apiHandler.HandlePause)
	http.HandleFunc("/resume", apiHandler.HandleResume)
	http.HandleFunc("/monitor", apiHandler.HandleMonitor)

	srv := &http.Server{Addr: ":8080"}
	log.Println("HTTP control API listening on :8080")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http serve: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("shutting down")

	if err := app.StopWorker(); err != nil {
		log.Printf("worker stop failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown failed: %v", err)
	}
	if err := asynqClient.Close(); err != nil {
		log.Printf("asynq client close failed: %v", err)
	}
	if err := redisClient.Close(); err != nil {
		log.Printf("redis client close failed: %v", err)
	}
}
