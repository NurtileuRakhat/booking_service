package postgres

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"github.com/jmoiron/sqlx"
)

type WorkspaceRepository struct {
	db *sqlx.DB
}

func NewWorkspaceRepository(db *sqlx.DB) repository.WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) GetWorkspaceByID(ctx context.Context, id int64) (*entity.Workspace, error) {
	var w entity.Workspace
	logger.Info("Getting workspace by id: %d", id)
	err := r.db.GetContext(ctx, &w, "SELECT * FROM workspaces WHERE id = $1", id)
	if err != nil {
		logger.Error("Failed to get workspace by id: %d, err: %v", id, err)
		return nil, err
	}
	logger.Info("Workspace found: id %d", id)
	return &w, nil
}

func (r *WorkspaceRepository) ListWorkspaces(ctx context.Context) ([]entity.Workspace, error) {
	var workspaces []entity.Workspace
	logger.Info("Selecting all workspaces")
	err := r.db.SelectContext(ctx, &workspaces, "SELECT * FROM workspaces ORDER BY id")
	if err != nil {
		logger.Error("Failed to select all workspaces, err: %v", err)
	}
	return workspaces, err
}

func (r *WorkspaceRepository) CreateWorkspace(ctx context.Context, workspace *entity.Workspace) (int64, error) {
	query := `INSERT INTO workspaces (name, type, hourly_rate, description, capacity) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int64
	logger.Info("Creating workspace: %s", workspace.Name)
	err := r.db.QueryRowContext(ctx, query, workspace.Name, workspace.Type, workspace.HourlyRate, workspace.Description, workspace.Capacity).Scan(&id)
	if err != nil {
		logger.Error("Failed to create workspace: %s, err: %v", workspace.Name, err)
		return 0, err
	}
	logger.Info("Workspace created successfully (id: %d, name: %s)", id, workspace.Name)
	return id, nil
}
