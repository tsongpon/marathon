package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tsongpon/marathon/internal/handler"
	"github.com/tsongpon/marathon/internal/repository"
	"github.com/tsongpon/marathon/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx := context.Background()

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		log.Fatal("SLACK_WEBHOOK_URL environment variable is required")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	firestoreClient, err := firestore.NewClientWithDatabase(ctx, projectID, "marathon")
	if err != nil {
		log.Fatalf("failed to create firestore client: %v", err)
	}
	defer firestoreClient.Close()
	onCallCalendarID := os.Getenv("ON_CALL_CALENDAR_ID")
	googleCalendarCredential := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	onCallRepo := repository.NewOnCallGoogleCalendarRepository(googleCalendarCredential, onCallCalendarID)

	alertRepo := repository.NewAlertFirestoreRepository(firestoreClient)
	notificationRepo := repository.NewSlackNotificationRepository()

	alertService := service.NewAlertService(alertRepo, notificationRepo, onCallRepo, slackWebhookURL)
	alertHandler := handler.NewAlertHttpHandler(alertService)

	e := echo.New()
	e.Use(middleware.RequestLogger())

	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.POST("/generic/alerts", alertHandler.CreateGenericAlerts)
	e.POST("/notify/alerts", alertHandler.NotifyAlerts)
	e.POST("/signoz/alerts", alertHandler.CreateSignozAlert)
	e.GET("/alerts/:id/ack", alertHandler.AckAlerts)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
