package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type Publisher interface {
	Publish(topic string, payload []byte, retain bool, qos byte) error
}

type TaskLogWriter interface {
	Create(ctx context.Context, log *model.ScheduledTaskLog) error
}

type Scheduler struct {
	taskRepo  *repository.ScheduledTaskRepository
	publisher Publisher
	logRepo   TaskLogWriter
	parser    cron.Parser
	stop      chan struct{}
	wg        sync.WaitGroup
}

func New(taskRepo *repository.ScheduledTaskRepository, publisher Publisher, logRepo TaskLogWriter) *Scheduler {
	return &Scheduler{
		taskRepo:  taskRepo,
		publisher: publisher,
		logRepo:   logRepo,
		parser:    cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		stop:      make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	logger.Log.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
	logger.Log.Info("Scheduler stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.tick(now)
		}
	}
}

func (s *Scheduler) tick(now time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tasks, err := s.taskRepo.ListAllEnabled(ctx)
	if err != nil {
		logger.Log.Errorf("Scheduler: list tasks failed: %v", err)
		return
	}

	nowMinute := now.Truncate(time.Minute)
	prevMinute := nowMinute.Add(-time.Minute)

	for _, task := range tasks {
		sched, err := s.parser.Parse(task.CronExpr)
		if err != nil {
			continue
		}
		next := sched.Next(prevMinute)
		if next.Before(nowMinute) || !next.Before(nowMinute.Add(time.Minute)) {
			continue
		}
		s.executeTask(ctx, task)
	}
}

func (s *Scheduler) executeTask(ctx context.Context, task *model.ScheduledTask) {
	var topic string
	var payload []byte

	switch task.ActionType {
	case "property_set":
		topic = fmt.Sprintf("devices/%s/telemetry/down", task.DeviceID)
		data := map[string]interface{}{task.PropertyID: json.RawMessage(task.Value)}
		payload, _ = json.Marshal(data)
	case "service_invoke":
		topic = fmt.Sprintf("devices/%s/service/invoke", task.DeviceID)
		msg := map[string]interface{}{
			"id":      fmt.Sprintf("sched_%s_%d", task.ID[:8], time.Now().Unix()),
			"service": task.ServiceID,
			"params":  json.RawMessage(task.Value),
		}
		payload, _ = json.Marshal(msg)
	default:
		return
	}

	publishErr := s.publisher.Publish(topic, payload, false, 1)

	// 写执行日志
	if s.logRepo != nil {
		entry := &model.ScheduledTaskLog{
			TaskID:     task.ID,
			UserID:     task.UserID,
			DeviceID:   task.DeviceID,
			DeviceName: task.DeviceName,
			TaskName:   task.Name,
			ActionType: task.ActionType,
			Topic:      topic,
			Payload:    json.RawMessage(payload),
			Status:     "success",
		}
		if publishErr != nil {
			entry.Status = "failed"
			entry.Error = publishErr.Error()
		}
		if err := s.logRepo.Create(ctx, entry); err != nil {
			logger.Log.Errorf("Scheduler: write task log failed: %v", err)
		}
	}

	if publishErr != nil {
		logger.Log.Errorf("Scheduler: publish failed for task %s: %v", task.ID, publishErr)
		return
	}

	logger.Log.Infof("Scheduler: executed task %s (%s) on device %s", task.Name, task.ActionType, task.DeviceID)
	_ = s.taskRepo.UpdateLastRunAt(ctx, task.ID)
}
