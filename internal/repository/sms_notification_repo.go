package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/tsongpon/marathon/internal/logger"
	"github.com/tsongpon/marathon/internal/model"
	"go.uber.org/zap"
)

type SMSNotificationRepository struct {
	snsClient *sns.Client
}

func NewSMSNotificationRepository() (*SMSNotificationRepository, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return &SMSNotificationRepository{
		snsClient: sns.NewFromConfig(cfg),
	}, nil
}

func (r *SMSNotificationRepository) SendNotification(ctx context.Context, alert model.Alert, onCall model.OnCall) error {
	ackMessage := fmt.Sprintf("To ack: https://marathon-443460999135.asia-southeast1.run.app/alerts/%s/ack", alert.ID)
	message := fmt.Sprintf("[ALERT] %s\n%s\n\n%s", alert.Title, alert.Details, ackMessage)

	input := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(onCall.PhoneNumber),
	}

	_, err := r.snsClient.Publish(ctx, input)
	if err != nil {
		logger.Error("Failed to send SMS", zap.String("phone", onCall.PhoneNumber), zap.Error(err))
		return fmt.Errorf("failed to send SMS to %s: %w", onCall.PhoneNumber, err)
	}

	log.Printf("SMS sent to oncall %s at %s for alert %s", onCall.Name, onCall.PhoneNumber, alert.ID)
	return nil
}
