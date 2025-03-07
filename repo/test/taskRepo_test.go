package test

import (
	"context"
	"database/sql"
	"jwt2/models"
	repo "jwt2/repo/src"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockTasks = []models.Task{
	{ID: 1, Title: "Task 1", UserID: "123"},
	{ID: 2, Title: "Task 2", UserID: "123"},
	{ID: 3, Title: "Task 3", UserID: "123"},
	{ID: 4, Title: "Task 4", UserID: "123"},
	{ID: 5, Title: "Task 5", UserID: "123"},
	{ID: 6, Title: "Task but bad", UserID: "123"},
	{ID: 7, Title: "Task but worst", UserID: "123"},
	{ID: 8, Title: "Task idk", UserID: "123"},
	{ID: 9, Title: "Task", UserID: "123"},
	{ID: 10, Title: "Task hm", UserID: "123"},
	{ID: 11, Title: "Task ...", UserID: "123"},
	{ID: 12, Title: "Task a", UserID: "123"},
	{ID: 13, Title: "Task b", UserID: "123"},
	{ID: 14, Title: "Task d", UserID: "123"},
	{ID: 15, Title: "Task c", UserID: "123"},
}

// Mock cho DB operations
type MockBunDB struct {
	mock.Mock
}

func (m *MockBunDB) NewSelect() repo.SelectQuerier {
	args := m.Called()
	return args.Get(0).(repo.SelectQuerier)
}

func (m *MockBunDB) NewInsert() repo.InsertQuerier {
	args := m.Called()
	return args.Get(0).(repo.InsertQuerier)
}
func (m *MockBunDB) NewDelete() repo.DeleteQuerier {
	args := m.Called()
	return args.Get(0).(repo.DeleteQuerier)
}
func (m *MockBunDB) NewUpdate() repo.UpdateQuerier {
	args := m.Called()
	return args.Get(0).(repo.UpdateQuerier)
}

// Mock cho Select Query
type MockSelectQuery struct {
	mock.Mock
}

func (m *MockSelectQuery) Model(model interface{}) repo.SelectQuerier {
	m.Called(model)
	switch v := model.(type) {
	case *[]models.Task:
		*v = mockTasks
	case *models.Task:
		*v = models.Task{ID: 1, Title: "Task 1", UserID: "123"}
	}
	return m
}

func (m *MockSelectQuery) Limit(limit int) repo.SelectQuerier {
	m.Called(limit)
	return m
}

func (m *MockSelectQuery) Offset(offSet int) repo.SelectQuerier {
	m.Called(offSet)
	return m
}

func (m *MockSelectQuery) Order(order string) repo.SelectQuerier {
	m.Called(order)
	return m
}
func (m *MockSelectQuery) Where(query string, args ...interface{}) repo.SelectQuerier {
	m.Called(query, args[0])
	return m
}

func (m *MockSelectQuery) Scan(ctx context.Context, dest interface{}) error {
	args := m.Called(ctx, dest)
	return args.Error(0)
}

// Mock cho Insert Query
type MockInsertQuery struct {
	mock.Mock
}

func (m *MockInsertQuery) Model(model interface{}) repo.InsertQuerier {
	m.Called(model)
	return m
}

func (m *MockInsertQuery) Exec(ctx context.Context) (sql.Result, error) {
	args := m.Called(ctx)
	return nil, args.Error(1)
}

func TestTaskRepo_CreateTask(t *testing.T) {
	// Setup
	mockDB := new(MockBunDB)
	mockQuery := new(MockInsertQuery)
	taskRepo := repo.NewTaskRepo(mockDB)
	ctx := context.Background()
	now := time.Now()

	task := &models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      1,
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      "123",
	}

	// Set expectations
	mockDB.On("NewInsert").Return(mockQuery)
	mockQuery.On("Model", task).Return(mockQuery)
	mockQuery.On("Exec", ctx).Return(nil, nil)

	// Execute
	err := taskRepo.CreateTask(ctx, task)

	// Assert
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestTaskRepo_GetAllTask(t *testing.T) {
	// Setup
	mockDB := new(MockBunDB)
	mockQuery := new(MockSelectQuery)
	taskRepo := repo.NewTaskRepo(mockDB)
	ctx := context.Background()
	userID := "123"
	page := &models.Paging{
		Page:      1,
		Limit:     5,
		Title:     "",
		StartDate: "",
		EndDate:   "",
		Status:    nil,
	}

	// Set expectations
	mockDB.On("NewSelect").Return(mockQuery)
	mockQuery.On("Model", mock.AnythingOfType("*[]models.Task")).Run(func(args mock.Arguments) {
		taskList := args.Get(0).(*[]models.Task)
		*taskList = mockTasks
	}).Return(mockQuery)
	mockQuery.On("Where", "user_id =?", userID).Return(mockQuery)

	mockQuery.On("Limit", page.Limit).Return(mockQuery)
	mockQuery.On("Offset", page.OffSet()).Return(mockQuery)
	mockQuery.On("Order", "id ASC").Return(mockQuery)

	mockQuery.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		//ctx := args.Get(0).(context.Context)
		taskList := args.Get(1).(*[]models.Task)
		offset := page.OffSet()
		limit := page.Limit
		total := len(mockTasks)
		if offset >= total {
			*taskList = []models.Task{}
			return
		}
		end := offset + limit
		if end > total {
			end = total
		}
		*taskList = mockTasks[offset:end]
	}).Return(nil)

	// Execute
	tasks, err := taskRepo.GetAllTask(ctx, userID, page)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, tasks, 5)
	assert.Equal(t, "Task 1", tasks[0].Title)
	assert.Equal(t, "Task 2", tasks[1].Title)
	mockDB.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestTaskRepo_GetTaskByID(t *testing.T) {
	// Setup
	mockDB := new(MockBunDB)
	mockQuery := new(MockSelectQuery)
	taskRepo := repo.NewTaskRepo(mockDB)
	ctx := context.Background()
	taskID := int64(1)

	// Set expectations
	mockDB.On("NewSelect").Return(mockQuery)
	mockQuery.On("Model", mock.AnythingOfType("*models.Task")).Return(mockQuery)
	mockQuery.On("Where", "id =?", taskID).Return(mockQuery)
	mockQuery.On("Scan", ctx).Return(nil)

	// Execute
	task, err := taskRepo.GetTaskByID(ctx, taskID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "Task 1", task.Title)
	mockDB.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}
