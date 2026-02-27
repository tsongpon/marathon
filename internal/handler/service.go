package handler

import (
	"context"
	"time"

	"github.com/tsongpon/marathon/internal/model"
)

type AlertService interface {
	CreateAlert(ctx context.Context, alert model.Alert) (model.Alert, error)
	DeleteAlert(ctx context.Context, id string) error
	Notify(ctx context.Context, asOf time.Time) error
}
