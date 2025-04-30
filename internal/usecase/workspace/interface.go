package workspace

import (
	"booking/internal/entity"
	"context"
)

type Service interface {
	GetWorkspaceByID(ctx context.Context, id int64) (*entity.Workspace, error)
	ListWorkspaces(ctx context.Context) ([]entity.Workspace, error)
	CreateWorkspace(ctx context.Context, workspace *entity.Workspace) (int64, error)
}
