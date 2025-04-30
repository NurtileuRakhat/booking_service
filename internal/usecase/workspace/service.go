package workspace

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"context"
)

type WorkspaceService struct {
	repo repository.WorkspaceRepository
}

func NewWorkspaceService(repo repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

func (s *WorkspaceService) GetWorkspaceByID(ctx context.Context, id int64) (*entity.Workspace, error) {
	return s.repo.GetWorkspaceByID(ctx, id)
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context) ([]entity.Workspace, error) {
	return s.repo.ListWorkspaces(ctx)
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, workspace *entity.Workspace) (int64, error) {
	return s.repo.CreateWorkspace(ctx, workspace)
}
