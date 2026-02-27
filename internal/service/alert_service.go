package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tsongpon/marathon/internal/logger"
	"github.com/tsongpon/marathon/internal/model"
	"go.uber.org/zap"
)

type AlertService struct {
	alertRepo        AlertRepository
	onCallRepo       OnCallRepository
	notificationRepo NotificationRepository
}

func NewAlertService(alertRepo AlertRepository, onCallRepo OnCallRepository, notificationRepo NotificationRepository) *AlertService {
	return &AlertService{
		alertRepo:        alertRepo,
		onCallRepo:       onCallRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *AlertService) CreateAlert(ctx context.Context, alert model.Alert) (model.Alert, error) {
	alert.CreatedAt = time.Now()
	alert.ID = uuid.New().String()
	err := s.alertRepo.CreateAlert(ctx, alert)
	if err != nil {
		return model.Alert{}, err
	}
	s.Notify(ctx, time.Now())
	return alert, nil
}

func (s *AlertService) DeleteAlert(ctx context.Context, id string) error {
	err := s.alertRepo.DeleteAlert(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *AlertService) Notify(ctx context.Context, asOf time.Time) error {
	alerts, err := s.alertRepo.GetAlerts(ctx)
	if err != nil {
		logger.Error("Failed to get alerts", zap.Error(err))
		return err
	}
	logger.Info("Number of alerts", zap.Int("alert", len(alerts)))
	if len(alerts) > 0 {
		onCalls, err := s.onCallRepo.GetOnCalls(ctx, asOf)
		if err != nil {
			return err
		}
		logger.Info("Number of oncalls", zap.Int("oncall", len(onCalls)))
		var wg sync.WaitGroup
		errCh := make(chan error, len(alerts)*len(onCalls))

		for _, alert := range alerts {
			for _, onCall := range onCalls {
				wg.Add(1)
				go func(alert model.Alert, onCall model.OnCall) {
					defer wg.Done()
					for i := range 5 {
						logger.Info("Sending alert to oncall", zap.String("phone", onCall.PhoneNumber), zap.Int("attempt", i+1))
						err := s.notificationRepo.SendNotification(ctx, alert, onCall)
						if err != nil {
							logger.Error("Fail to send alert notification", zap.Error(err))
							errCh <- err
							return
						}
						if i < 4 {
							time.Sleep(10 * time.Second)
						}
					}
				}(alert, onCall)
			}
		}

		wg.Wait()
		close(errCh)

		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}
