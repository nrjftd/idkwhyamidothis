package test

import (
	"context"
	pagev1 "jwt2/gen/go/proto/page/v1"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/models"
	service "jwt2/service/src"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) CreateTask(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
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
func (m *MockTaskRepo) DeleteTaskByID(ctx context.Context, taskID int64) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}
func (m *MockTaskRepo) UpdateTaskByID(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

// server streaming
type mockTaskService_GetTaskServer struct {
	grpc.ServerStream
	mock.Mock
}

func (m *mockTaskService_GetTaskServer) Send(resp *v1.GetTaskResponse) error {
	args := m.Called(resp)
	return args.Error(0)
}

// test
func TestCreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	service := service.NewTaskService(mockRepo)
	req := &v1.CreateTaskRequest{
		Task: &v1.Task{
			Id:     1,
			Title:  "test task",
			UserID: "1",
		},
	}
	task := models.ProtoToModel(req.Task)
	mockRepo.On("CreateTask", mock.Anything, task).Return(nil)
	resp, err := service.CreateTask(context.Background(), req)
	assert.Equal(t, req.Task.Title, resp.Task.Title)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTaskByID(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	service := service.NewTaskService(mockRepo)
	task := &models.Task{ID: 1, Title: "task 1", UserID: "1"}
	req := &v1.GetTaskByIdRequest{Id: 1}
	mockRepo.On("GetTaskByID", mock.Anything, int64(1)).Return(task, nil)
	result, err := service.GetTaskByID(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, task.ID, int64(result.Task.Id))
	assert.Equal(t, task.Title, result.Task.Title)
	mockRepo.AssertExpectations(t)

}

//==================================================================//

// func TestGetAllTask(t *testing.T) {
// 	mockRepo := new(MockTaskRepo)
// 	service := service.NewTaskService(mockRepo)
// 	tasks := []models.Task{
// {ID: 1, Title: "task 1", UserID: "1"},
// {ID: 2, Title: "task but second", UserID: "2"},
// {ID: 3, Title: "task 2", UserID: "1"},
// {ID: 4, Title: "task but ", UserID: "1"},
// {ID: 5, Title: "task 132", UserID: "1"},
// {ID: 6, Title: "task 1 second", UserID: "1"},
// {ID: 7, Title: "task 11 aa", UserID: "1"},
// {ID: 8, Title: "task 1 aa", UserID: "1"},
// {ID: 9, Title: "task 1a w", UserID: "1"},
// {ID: 10, Title: "task 1 11", UserID: "1"},
// 	}
// 	mockRepo.On("GetAllTask", mock.Anything, "1", mock.Anything).Return(tasks, nil)

// 	mockStream := new(mockTaskService_GetTaskServer)
// 	req := &v1.GetTaskRequest{UserID: "1",
// 		Page: &pagev1.PageRequest{
// 			Page:  1,
// 			Limit: 5},
// 	}

// 	for i, task := range tasks[:5] {
// 		expectedTask := task.ToProto()
// 		fmt.Printf("expectedTask %d: %+v\n", i+1, expectedTask)
// 		mockStream.On("Send", mock.MatchedBy(func(resp *v1.GetTaskResponse) bool {
// 			return proto.Equal(resp.Task, expectedTask)
// 		})).Return(nil)

// 	}
// 	err := service.GetAllTask(context.Background(), req, mockStream)
// 	assert.NoError(t, err)
// 	mockStream.AssertNumberOfCalls(t, "Send", 5)
// 	mockRepo.AssertExpectations(t)
// }

func TestGetAllTask(t *testing.T) {
	//repo
	mockRepo := new(MockTaskRepo)
	service := service.NewTaskService(mockRepo)
	//mock data
	tasks := []models.Task{
		{ID: 1, Title: "task 1", UserID: "1"},
		{ID: 2, Title: "task but second", UserID: "1"},
		{ID: 3, Title: "task 2", UserID: "1"},
		{ID: 4, Title: "task but ", UserID: "1"},
		{ID: 5, Title: "task 132", UserID: "1"},
		{ID: 6, Title: "task 1 second", UserID: "1"},
		{ID: 7, Title: "task 11 aa", UserID: "1"},
		{ID: 8, Title: "task 1 aa", UserID: "1"},
		{ID: 9, Title: "task 1a w", UserID: "1"},
		{ID: 10, Title: "task 1 11", UserID: "1"},
		{ID: 11, Title: "task 132", UserID: "1"},
		{ID: 12, Title: "task 1 second", UserID: "1"},
		{ID: 13, Title: "task 11 aa", UserID: "1"},
		{ID: 14, Title: "task 1 aa", UserID: "1"},
		{ID: 15, Title: "task 1a w", UserID: "1"},
	}
	//get task
	taskPage1 := tasks[:5]
	taskPage2 := tasks[5:10]
	taskPage3 := tasks[10:15]

	//page request
	pageRequest := &pagev1.PageRequest{
		Page:      1,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	pageRequest2 := &pagev1.PageRequest{
		Page:      2,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	pageRequest3 := &pagev1.PageRequest{
		Page:      3,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}
	//get task request
	req := &v1.GetTaskRequest{
		UserID: "123",
		Page:   pageRequest,
	}
	req2 := &v1.GetTaskRequest{
		UserID: "123",
		Page:   pageRequest2,
	}
	req3 := &v1.GetTaskRequest{
		UserID: "123",
		Page:   pageRequest3,
	}

	//
	pageModel := &models.Paging{
		Page:      1,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	pageModel2 := &models.Paging{
		Page:      2,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}

	pageModel3 := &models.Paging{
		Page:      3,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    []string{},
	}
	mockRepo.On("GetAllTask", mock.Anything, "123", pageModel).Return(taskPage1, nil)
	mockRepo.On("GetAllTask", mock.Anything, "123", pageModel2).Return(taskPage2, nil)
	mockRepo.On("GetAllTask", mock.Anything, "123", pageModel3).Return(taskPage3, nil)
	//stream
	stream := new(mockTaskService_GetTaskServer)
	stream.On("Send", mock.Anything).Return(nil)

	//test
	err := service.GetAllTask(context.Background(), req, stream)
	assert.NoError(t, err)

	err = service.GetAllTask(context.Background(), req2, stream)
	assert.NoError(t, err)

	err = service.GetAllTask(context.Background(), req3, stream)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	stream.AssertExpectations(t)
}
