package repository

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/tsongpon/marathon/internal/model"
	"google.golang.org/api/iterator"
)

const alertsCollection = "alerts"

type AlertFirestoreRepository struct {
	client *firestore.Client
}

func NewAlertFirestoreRepository(client *firestore.Client) *AlertFirestoreRepository {
	return &AlertFirestoreRepository{client: client}
}

func (r *AlertFirestoreRepository) CreateAlert(ctx context.Context, alert model.Alert) error {
	_, err := r.client.Collection(alertsCollection).Doc(alert.ID).Set(ctx, map[string]any{
		"id":              alert.ID,
		"title":           alert.Title,
		"details":         alert.Details,
		"created_at":      alert.CreatedAt,
		"is_acknowledged": alert.IsAcknowledged,
	})
	return err
}

func (r *AlertFirestoreRepository) DeleteAlert(ctx context.Context, id string) error {
	_, err := r.client.Collection(alertsCollection).Doc(id).Delete(ctx)
	return err
}

func (r *AlertFirestoreRepository) GetAlerts(ctx context.Context) ([]model.Alert, error) {
	var alerts []model.Alert
	iter := r.client.Collection(alertsCollection).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var alert model.Alert
		if err := doc.DataTo(&alert); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}
