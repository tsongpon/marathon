package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/tsongpon/marathon/internal/model"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type OnCallGoogleCalendarRepository struct {
	base64GoogleCalendarCredential string
	calendarID                     string
}

func NewOnCallGoogleCalendarRepository(base64GoogleCalendarCredential string, calendarID string) *OnCallGoogleCalendarRepository {
	return &OnCallGoogleCalendarRepository{
		base64GoogleCalendarCredential: base64GoogleCalendarCredential,
		calendarID:                     calendarID,
	}
}

func (r *OnCallGoogleCalendarRepository) GetOnCalls(ctx context.Context, asOf time.Time) ([]model.OnCall, error) {
	credential, err := base64.StdEncoding.DecodeString(r.base64GoogleCalendarCredential)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 credential: %v", err)
	}

	config, err := google.JWTConfigFromJSON(credential, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := config.Client(ctx)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	beginningOfDay := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 9, 0, 0, 0, asOf.Location()).Format(time.RFC3339)
	endOfDay := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), asOf.Location()).Format(time.RFC3339)
	log.Printf("Get event of : %s, from calendar : %s", asOf.Format(time.DateOnly), r.calendarID)

	todayOncallEvents, err := srv.Events.List(r.calendarID).ShowDeleted(false).
		SingleEvents(true).TimeMin(beginningOfDay).TimeMax(endOfDay).MaxResults(50).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	var onCalls []model.OnCall
	for _, item := range todayOncallEvents.Items {
		onCall := model.OnCall{Name: item.Summary}
		onCalls = append(onCalls, onCall)
	}
	return onCalls, nil
}
