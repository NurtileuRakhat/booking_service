package workspace

import (
	"context"
	"errors"
	"testing"

	"booking/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWorkspaceRepo struct{ mock.Mock }

func (m *MockWorkspaceRepo) GetWorkspaceByID(ctx context.Context, id int64) (*entity.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Workspace), args.Error(1)
}
func (m *MockWorkspaceRepo) ListWorkspaces(ctx context.Context) ([]entity.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Workspace), args.Error(1)
}
func (m *MockWorkspaceRepo) CreateWorkspace(ctx context.Context, workspace *entity.Workspace) (int64, error) {
	args := m.Called(ctx, workspace)
	return args.Get(0).(int64), args.Error(1)
}

func TestGetWorkspaceByID_Success(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	ws := &entity.Workspace{ID: 1, Name: "Room1"}
	repo.On("GetWorkspaceByID", ctx, int64(1)).Return(ws, nil)
	result, err := service.GetWorkspaceByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, ws, result)
	repo.AssertExpectations(t)
}

func TestGetWorkspaceByID_NotFound(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	repo.On("GetWorkspaceByID", ctx, int64(2)).Return(nil, errors.New("not found"))
	result, err := service.GetWorkspaceByID(ctx, 2)
	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestListWorkspaces_Success(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	wsList := []entity.Workspace{{ID: 1, Name: "Room1"}, {ID: 2, Name: "Room2"}}
	repo.On("ListWorkspaces", ctx).Return(wsList, nil)
	result, err := service.ListWorkspaces(ctx)
	assert.NoError(t, err)
	assert.Equal(t, wsList, result)
	repo.AssertExpectations(t)
}

func TestListWorkspaces_Error(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	repo.On("ListWorkspaces", ctx).Return(nil, errors.New("db error"))
	result, err := service.ListWorkspaces(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestCreateWorkspace_Success(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	ws := &entity.Workspace{Name: "Room1"}
	repo.On("CreateWorkspace", ctx, ws).Return(int64(10), nil)
	id, err := service.CreateWorkspace(ctx, ws)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), id)
	repo.AssertExpectations(t)
}

func TestCreateWorkspace_Error(t *testing.T) {
	repo := new(MockWorkspaceRepo)
	service := NewWorkspaceService(repo)
	ctx := context.Background()
	ws := &entity.Workspace{Name: "Room1"}
	repo.On("CreateWorkspace", ctx, ws).Return(int64(0), errors.New("insert error"))
	id, err := service.CreateWorkspace(ctx, ws)
	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
	repo.AssertExpectations(t)
}
