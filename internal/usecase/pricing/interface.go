package pricing

import (
	"context"
	"time"
)

type Service interface {
	CalculatePrice(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error)
}
