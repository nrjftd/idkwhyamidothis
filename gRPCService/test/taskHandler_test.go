package test

import (
	"context"
	"errors"
	gRPCSerivce "jwt2/gRPCService/src"
	"jwt2/models"
	pagev1 "jwt2/proto/page/v1"
	v1 "jwt2/proto/task/v1"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockTaskRepo struct {
	mock.Mock
}
type mockTaskService_GetTaskServer struct {
	grpc.ServerStream
	mock.Mock
}
type MockRedisClient struct {
	mock.Mock
	//*redis.Client
}

func (m *MockRedisClient) Pub(ctx context.Context, channel string, message interface{}) error {
	args := m.Called(ctx, channel, message)
	return args.Error(0)
}
func (m *MockRedisClient) Sub(ctx context.Context, channel string) (<-chan string, error) {
	args := m.Called(ctx, channel)
	msgChan := make(chan string, 1)
	if args.Get(0) != nil {
		msgChan <- args.Get(0).(string)
	}
	return msgChan, args.Error(0)

}
func (m *MockRedisClient) Del(ctx context.Context, key ...string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}
func (m *MockTaskRepo) GetAllTask(ctx context.Context, userID string, page *models.Paging) ([]models.Task, error) {
	args := m.Called(ctx, userID, page)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepo) GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error) {
	args := m.Called(ctx, taskID)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepo) CreateTask(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}
func (m *MockTaskRepo) DeleteTaskByID(ctx context.Context, taskID int64) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}
func (m *MockTaskRepo) UpdateTaskByID(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}
func (m *mockTaskService_GetTaskServer) Send(resp *v1.GetTaskResponse) error {
	args := m.Called(resp)
	return args.Error(0)
}
func (m *mockTaskService_GetTaskServer) Context() context.Context {
	return context.WithValue(context.Background(), "user_id", "user1")
}
func TestGetTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)

	tasks := []models.Task{
		{ID: 1, Title: "Task 1", Description: "Description 1", UserID: "user1"},
		{ID: 2, Title: "Task 2", Description: "Description 2", UserID: "user1"},
		{ID: 3, Title: "Task 3", Description: "Description 3", UserID: "user1"},
	}

	pageRequest := &pagev1.PageRequest{
		Page:      1,
		Limit:     10,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	req := &v1.GetTaskRequest{
		UserID: "user1",
		Page:   pageRequest,
	}

	pageModel := &models.Paging{
		Page:      1,
		Limit:     10,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	mockRepo.On("GetAllTask", mock.Anything, "user1", pageModel).Return(tasks, nil)

	stream := new(mockTaskService_GetTaskServer)
	stream.On("Send", mock.Anything).Return(nil)

	err := handler.GetTask(req, stream)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	stream.AssertExpectations(t)
}

//========================UNDONE========================//

func TestGetTaskByID(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	task := &models.Task{ID: 1, Title: "Task 1", Description: "Description 1", UserID: "user1"}
	req := &v1.GetTaskByIdRequest{Id: 1}
	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(task, nil)
	result, err := handler.GetTaskById(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, task.ID, int64(result.Task.Id))
	assert.Equal(t, task.Title, result.Task.Title)
	mockRepo.AssertExpectations(t)

}

func TestCreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	req := &v1.CreateTaskRequest{
		Task: &v1.Task{
			Id:     1,
			Title:  "Task 1",
			UserID: "user1",
		},
	}
	task := models.ProtoToModel(req.Task)
	mockRepo.On("CreateTask", mock.Anything, task).Return(nil)
	resp, err := handler.CreateTask(context.Background(), req)
	assert.Equal(t, req.Task.Title, resp.Task.Title)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateTaskBadRequest(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	req := &v1.CreateTaskRequest{
		Task: nil,
	}
	resp, err := handler.CreateTask(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestDeleteTaskByID(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	mockRedis := new(MockRedisClient)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, mockRedis)
	task := &models.Task{ID: 1, Title: "Task 1", Description: "Description 1", UserID: "user1"}
	req := &v1.DeleteTaskByIdRequest{Id: 1}
	ctx := context.WithValue(context.Background(), "user_id", "user1")
	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(task, nil)
	mockRepo.On("DeleteTaskByID", mock.Anything, int64(1)).Return(nil)
	mockRedis.On("Pub", mock.Anything, "task_event", "task_deleted|1").Return(nil)
	mockRedis.On("Del", mock.Anything, mock.Anything).Return(nil)

	result, err := handler.DeleteTaskByID(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Task deleted successfully", result.Message)

	mockRepo.AssertExpectations(t)
}

func TestDeleteTaskByIDNotFound(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	req := &v1.DeleteTaskByIdRequest{Id: 1}

	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(nil, errors.New("task not found"))

	result, err := handler.DeleteTaskByID(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "task with ID 1 not found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestDeleteTaskByIDForbidden(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	task := &models.Task{ID: 1, Title: "Task 1", Description: "Description 1", UserID: "user2"}
	req := &v1.DeleteTaskByIdRequest{Id: 1}

	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(task, nil)

	result, err := handler.DeleteTaskByID(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Forbidden", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestGetTasByID(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	handler := gRPCSerivce.NewTaskHandler(mockRepo, nil)
	tasks := []models.Task{
		{ID: 1, Title: "Task 1", Description: "Description 1", UserID: "user1"},
		{ID: 2, Title: "Task 2", Description: "Description 2", UserID: "user1"},
		{ID: 3, Title: "Task 3", Description: "Description 3", UserID: "user1"},
	}
	req := &v1.GetTaskByIdRequest{Id: 1}
	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(&tasks[0], nil)
	result, err := handler.GetTaskById(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, tasks[0].ID, int64(result.Task.Id))
}
