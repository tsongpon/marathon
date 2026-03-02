package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tsongpon/marathon/internal/logger"
	"github.com/tsongpon/marathon/internal/model"
	"go.uber.org/zap"
)

type AlertService struct {
	alertRepo        AlertRepository
	notificationRepo NotificationRepository
	onCallRepo       OnCallRepository
	slackWebhookURL  string
}

func NewAlertService(alertRepo AlertRepository, notificationRepo NotificationRepository, onCallRepo OnCallRepository, slackWebhookURL string) *AlertService {
	return &AlertService{
		alertRepo:        alertRepo,
		notificationRepo: notificationRepo,
		onCallRepo:       onCallRepo,
		slackWebhookURL:  slackWebhookURL,
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
		onDutyOnCall, err := s.onCallRepo.GetOnCalls(ctx, asOf)
		if err != nil {
			logger.Error("Failed to get oncall", zap.Error(err))
			return err
		}
		var onCallNames strings.Builder
		for i, onCall := range onDutyOnCall {
			onCallNames.WriteString(onCall.Name)
			if i < len(onDutyOnCall)-1 {
				onCallNames.WriteString("\n")
			}
		}
		onCall := model.OnCall{Name: onCallNames.String(), SlackWebhookURL: s.slackWebhookURL}
		for _, alert := range alerts {
			for i := range 5 {
				logger.Info("Sending alert to oncall", zap.String("SlackWebhookURL", onCall.SlackWebhookURL), zap.Int("attempt", i+1))
				err := s.notificationRepo.SendNotification(ctx, alert, onCall)
				if err != nil {
					logger.Error("Fail to send alert notification", zap.Error(err))
				}
				if i < 4 {
					time.Sleep(10 * time.Second)
				}
			}
		}
	}
	return nil
}
