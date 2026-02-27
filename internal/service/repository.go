package service

import (
	"context"
	"time"

	"github.com/tsongpon/marathon/internal/model"
)

type AlertRepository interface {
	CreateAlert(ctx context.Context, alert model.Alert) error
	DeleteAlert(ctx context.Context, id string) error
	GetAlerts(ctx context.Context) ([]model.Alert, error)
}

type NotificationRepository interface {
	SendNotification(ctx context.Context, alert model.Alert, onCall model.OnCall) error
}

type OnCallRepository interface {
	GetOnCalls(ctx context.Context, asOf time.Time) ([]model.OnCall, error)
}
