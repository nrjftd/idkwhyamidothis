package test

import (
	"bytes"
	"context"
	"encoding/json"
	controller "jwt2/controller/src"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/models"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, task *v1.CreateTaskRequest) (*v1.CreateTaskResponse, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*v1.CreateTaskResponse), args.Error(0)
}

func (m *MockTaskService) GetAllTask(ctx context.Context, userID string) (*v1.GetTaskResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*v1.GetTaskResponse), args.Error(1)
}

func (m *MockTaskService) GetTaskByID(ctx context.Context, ID *v1.GetTaskByIdRequest) (*v1.GetTaskByIdResponse, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(*v1.GetTaskByIdResponse), args.Error(1)
}
func (m *MockTaskService) DeleteTaskByID(ctx context.Context, ID *v1.DeleteTaskByIdRequest) (*v1.DeleteTaskByIdResponse, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(*v1.DeleteTaskByIdResponse), args.Error(1)
}

func (m *MockTaskService) UpdateTaskByID(ctx context.Context, task *v1.UpdateTaskRequest) (*v1.UpdateTaskResponse, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*v1.UpdateTaskResponse), args.Error(1)
}

// undone
func TestCreateTask(t *testing.T) {
	mockService := new(MockTaskService)
	controller := controller.NewTaskController(mockService)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/users/task", controller.CreateTask)

	task := models.Task{Title: "test Task"}
	protoTask := task.ToProto()
	req := &v1.CreateTaskRequest{Task: protoTask}
	mockService.On("CreateTask", mock.Anything, req).Return(&v1.CreateTaskResponse{Task: protoTask}, nil)

	jsonTask, _ := json.Marshal(task)
	hreq, _ := http.NewRequest(http.MethodPost, "/users/tasks", bytes.NewBuffer(jsonTask))
	hreq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = hreq
	ctx.Set("user_id", "1")

	controller.CreateTask(ctx)
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

// func TestGetAllTask(t *testing.T) {
// 	mockService := new(MockTaskService)
// 	controller := controller.NewTaskController(mockService)
// 	gin.SetMode(gin.TestMode)
// 	tasks := []models.Task{
// 		{ID: 1, Title: "task 1", UserID: "1"},
// 		{ID: 2, Title: "task but second", UserID: "1"},
// 		{ID: 3, Title: "task but third", UserID: "1"},
// 		{ID: 4, Title: "still task", UserID: "2"},
// 	}
// 	mockService.On("GetAllTask", mock.Anything, "1").Return(tasks, nil)
// 	req, _ := http.NewRequest(http.MethodGet, "/users/tasks", nil)
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)
// 	ctx.Request = req
// 	ctx.Set("user_id", "1")
// 	controller.GetAllTask(ctx)
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockService.AssertExpectations(t)
// }
// func TestGetTaskByID(t *testing.T) {
// 	mockService := new(MockTaskService)
// 	controller := controller.NewTaskController(mockService)
// 	gin.SetMode(gin.TestMode)
// 	task := &models.Task{ID: 1, Title: "task 1", UserID: "1"}
// 	mockService.On("GetTaskByID", mock.Anything, int64(1)).Return(task, nil)
// 	req, _ := http.NewRequest(http.MethodGet, "/users/task/1", nil)
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)
// 	ctx.Request = req
// 	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
// 	controller.GetTaskByID(ctx)
// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockService.AssertExpectations(t)
// }
