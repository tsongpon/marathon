package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tsongpon/marathon/internal/logger"
	"github.com/tsongpon/marathon/internal/model"
	"go.uber.org/zap"
)

type SlackNotificationRepository struct {
	httpClient *http.Client
}

func NewSlackNotificationRepository() *SlackNotificationRepository {
	return &SlackNotificationRepository{
		httpClient: &http.Client{},
	}
}

func (r *SlackNotificationRepository) SendNotification(ctx context.Context, alert model.Alert, onCall model.OnCall) error {
	payload := buildSlackPayload(alert, onCall)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, onCall.SlackWebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send Slack notification", zap.Error(err))
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Slack webhook returned non-200 status", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	logger.Info("Slack notification sent", zap.String("alert_id", alert.ID), zap.String("on_call", onCall.Name))
	return nil
}

func buildSlackPayload(alert model.Alert, onCall model.OnCall) map[string]any {
	ackURL := fmt.Sprintf("https://marathon-443460999135.asia-southeast1.run.app/alerts/%s/ack", alert.ID)

	return map[string]any{
		"text": fmt.Sprintf("🚨 On-Call Alert: %s", alert.Title),
		"blocks": []map[string]any{
			{
				"type": "header",
				"text": map[string]any{
					"type":  "plain_text",
					"text":  "🚨 On-Call Alert",
					"emoji": true,
				},
			},
			{
				"type": "section",
				"fields": []map[string]any{
					{"type": "mrkdwn", "text": "*Severity:*\n🔴 Critical"},
					{"type": "mrkdwn", "text": "*Status:*\n🔥 Firing"},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Service:*\n%s", alert.Title)},
					{"type": "mrkdwn", "text": "*Environment:*\nProduction"},
				},
			},
			{
				"type": "section",
				"text": map[string]any{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Alert:* %s\n*Description:* %s", alert.Title, alert.Details),
				},
			},
			{
				"type": "section",
				"fields": []map[string]any{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Started At:*\n%s", alert.CreatedAt.In(time.FixedZone("Asia/Bangkok", 7*60*60)).Format("2006-01-02 15:04:05"))},
					{"type": "mrkdwn", "text": fmt.Sprintf("*On-Call:*\n%s", onCall.Name)},
				},
			},
			{
				"type": "actions",
				"elements": []map[string]any{
					{
						"type": "button",
						"text": map[string]any{"type": "plain_text", "text": "📊 Signoz Dashboard", "emoji": true},
						"url":  "https://signoz.yourdomain.com/dashboard/prod-api",
					},
					{
						"type": "button",
						"text": map[string]any{"type": "plain_text", "text": "📖 Confluence Runbook", "emoji": true},
						"url":  "https://confluence.yourdomain.com/runbook/high-cpu",
					},
					{
						"type":  "button",
						"text":  map[string]any{"type": "plain_text", "text": "✅ Acknowledge", "emoji": true},
						"style": "primary",
						"url":   ackURL,
					},
				},
			},
			{
				"type": "divider",
			},
			{
				"type": "context",
				"elements": []map[string]any{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("🔔 Alert ID: `%s` | Source: SigNoz | <https://your-alert-system.com/alerts|View All Alerts>", alert.ID),
					},
				},
			},
		},
	}
}
