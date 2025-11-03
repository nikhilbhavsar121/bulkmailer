package worker

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"bulkmailer/internal/email"

	"github.com/hibiken/asynq"
)

func (w *Worker) StartWorker() error {
	w.serverMu.Lock()
	defer w.serverMu.Unlock()
	if w.asynqSrv != nil {
		return nil
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: w.RedisClient.Options().Addr},
		asynq.Config{
			Concurrency:     20,
			Queues:          map[string]int{QueueDefault: 1},
			ShutdownTimeout: 10 * time.Second,
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailSend, w.emailTaskHandler)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Printf("asynq server stopped: %v", err)
		}
	}()
	w.asynqSrv = srv
	log.Println("worker started")
	return nil
}

func (w *Worker) StopWorker() error {
	w.serverMu.Lock()
	defer w.serverMu.Unlock()
	if w.asynqSrv == nil {
		return nil
	}
	w.asynqSrv.Shutdown()
	w.asynqSrv = nil
	log.Println("worker stopped")
	return nil
}

func (w *Worker) emailTaskHandler(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("bad payload: %v", err)
		return err
	}
	if w.IsPaused(ctx) {
		return w.requeue(p, 5*time.Second)
	}
	if !w.RateLimiter.Allow(GlobalRateKey) {
		log.Printf("[RATE LIMIT] Too many emails for %s — delaying\n", GlobalRateKey)
		return w.requeue(p, 10*time.Second)
	}
	if err := email.SendEmail(p.To, p.Subject, p.Body); err != nil {
		w.statsMu.Lock()
		w.failedCount++
		w.statsMu.Unlock()
		log.Printf("send email failed to %s: %v", p.To, err)
		return err
	}
	w.statsMu.Lock()
	w.sentCount++
	w.statsMu.Unlock()
	return nil
}

func (w *Worker) requeue(p EmailPayload, delay time.Duration) error {
	nt, err := NewEmailTask(p)
	if err != nil {
		return fmt.Errorf("requeue task create failed: %w", err)
	}
	_, err = w.AsynqClient.Enqueue(nt, asynq.Queue(QueueDefault), asynq.ProcessIn(delay))
	if err != nil {
		return fmt.Errorf("requeue failed: %w", err)
	}
	return nil
}

func NewEmailTask(p EmailPayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailSend, b, asynq.MaxRetry(3)), nil
}

func (w *Worker) ParseCSVAndEnqueue(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	_, err = r.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	count := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("csv read error: %v", err)
			continue
		}
		if len(record) < 1 {
			continue
		}

		email := record[0]
		subject, body := "Hello", "Default body"
		if len(record) >= 2 && record[1] != "" {
			subject = record[1]
		}
		if len(record) >= 3 && record[2] != "" {
			body = record[2]
		}

		task, err := NewEmailTask(EmailPayload{To: email, Subject: subject, Body: body})
		if err != nil {
			log.Printf("task create failed for %s: %v", email, err)
			continue
		}
		_, err = w.AsynqClient.Enqueue(task, asynq.Queue(QueueDefault))
		if err != nil {
			log.Printf("enqueue err %v", err)
			continue
		}
		count++
		w.statsMu.Lock()
		w.enqueuedCount++
		w.statsMu.Unlock()
	}
	log.Printf("enqueued %d tasks from %s", count, path)
	return count, nil
}

func (w *Worker) GetMonitorInfo(ctx context.Context) map[string]interface{} {
	qinfo, err := w.Inspector.GetQueueInfo(QueueDefault)
	if err != nil {
		qinfo = nil
	}

	w.statsMu.Lock()
	sc := w.sentCount
	fc := w.failedCount
	ec := w.enqueuedCount
	w.statsMu.Unlock()

	return map[string]interface{}{
		"queue":      QueueDefault,
		"queue_info": qinfo,
		"sent":       sc,
		"failed":     fc,
		"enqueued":   ec,
		"paused":     w.IsPaused(ctx),
	}
}

func (w *Worker) IsPaused(ctx context.Context) bool {
	if w.RedisClient == nil {
		return false
	}
	v, err := w.RedisClient.Get(ctx, "email:paused").Result()
	if err != nil {
		return false
	}
	return v == "1"
}

func (w *Worker) PauseProcessing(ctx context.Context) error {
	return w.RedisClient.Set(ctx, "email:paused", "1", 0).Err()
}

func (w *Worker) ResumeProcessing(ctx context.Context) error {
	return w.RedisClient.Del(ctx, "email:paused").Err()
}

func (w *Worker) SetCSVPath(path string) {
	w.CSVPath = path
}

func (w *Worker) GetCSVPath() string {
	return w.CSVPath
}
