package pricing

import (
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"time"
)

type PricingService struct {
	workspaceRepo repository.WorkspaceRepository
}

func NewPricingService(workspaceRepo repository.WorkspaceRepository) *PricingService {
	return &PricingService{workspaceRepo: workspaceRepo}
}

func (s *PricingService) CalculatePrice(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
	workspace, err := s.workspaceRepo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Error getting workspace by ID %d: %v", workspaceID, err)
		return 0, err
	}
	dur := end.Sub(start).Hours()
	price := workspace.HourlyRate * dur

	// вечером +20%
	if start.Hour() >= 10 && end.Hour() <= 18 {
		price *= 1.2
	}
	return price, nil
}
