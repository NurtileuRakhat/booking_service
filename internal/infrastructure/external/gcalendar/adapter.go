package gcalendar

import (
	"context"
	"fmt"
	"booking/pkg/logger"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"os"
	"time"
)

type GCalendarAdapter struct {
	srv        *calendar.Service
	calendarID string
}

func NewGCalendarAdapter(credentialsPath, calendarID string) (*GCalendarAdapter, error) {
	logger.Info("Initializing Google Calendar Adapter with credentials: %s", credentialsPath)
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		logger.Error("Unable to read credentials: %v", err)
		return nil, fmt.Errorf("unable to read credentials: %v", err)
	}
	config, err := google.JWTConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		logger.Error("Unable to parse credentials: %v", err)
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx)))
	if err != nil {
		logger.Error("Unable to create calendar service: %v", err)
		return nil, fmt.Errorf("unable to create calendar service: %v", err)
	}
	logger.Info("Google Calendar Adapter initialized successfully for calendarID: %s", calendarID)
	return &GCalendarAdapter{srv: srv, calendarID: calendarID}, nil
}

func (g *GCalendarAdapter) ScheduleEvent(ctx context.Context, summary, description string, start, end time.Time) (string, error) {
	logger.Info("Scheduling event: %s | %s | %s - %s", summary, description, start.Format(time.RFC3339), end.Format(time.RFC3339))
	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start:       &calendar.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:         &calendar.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
	created, err := g.srv.Events.Insert(g.calendarID, event).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return created.Id, nil
}

func (g *GCalendarAdapter) DeleteEvent(ctx context.Context, eventID string) error {
	return g.srv.Events.Delete(g.calendarID, eventID).Context(ctx).Do()
}
