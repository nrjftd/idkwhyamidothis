package service

import (
	"context"
	"errors"
	"fmt"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/models"
	repo "jwt2/repo/src"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type taskService struct {
	repo repo.TaskRepository
}
type TaskService interface {
	CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskResponse, error)
	GetAllTask(ctx context.Context, req *v1.GetTaskRequest, stream v1.TaskService_GetTaskServer) error
	GetTaskByID(ctx context.Context, req *v1.GetTaskByIdRequest) (*v1.GetTaskByIdResponse, error)
	DeleteTaskByID(ctx context.Context, taskID *v1.DeleteTaskByIdRequest) (*v1.DeleteTaskByIdResponse, error)
	UpdateTaskByID(ctx context.Context, task *v1.UpdateTaskRequest) (*v1.UpdateTaskResponse, error)
}

func NewTaskService(repo repo.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (service *taskService) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskResponse, error) {
	if req.Task == nil {
		return nil, errors.New("task data is required") //avoid panic
	}
	task := models.ProtoToModel(req.Task)
	err := service.repo.CreateTask(ctx, task)
	if err != nil {
		fmt.Printf("Error creating task in repository: %v\n", err)
		return nil, err
	}
	return &v1.CreateTaskResponse{Task: task.ToProto()}, nil
}

func (service *taskService) GetAllTask(ctx context.Context, req *v1.GetTaskRequest, stream v1.TaskService_GetTaskServer) error {
	if req.UserID == "" {
		return status.Error(codes.InvalidArgument, "user id is required")
	}
	pageModel := &models.Paging{
		Page:      int(req.Page.Page),
		Limit:     int(req.Page.Limit),
		Title:     req.Page.Title,
		StartDate: req.Page.StartDate,
		EndDate:   req.Page.EndDate,
		Status:    req.Page.Status,
	}
	pageModel.Process()

	tasks, err := service.repo.GetAllTask(ctx, req.UserID, pageModel)

	if err != nil {
		return status.Errorf(codes.Internal, "failed to get task: %v", err)
	}
	for _, task := range tasks {
		err := stream.Send(&v1.GetTaskResponse{Task: task.ToProto()})
		if err != nil {
			return err
		}
	}
	return nil
}
func (service *taskService) GetTaskByID(ctx context.Context, req *v1.GetTaskByIdRequest) (*v1.GetTaskByIdResponse, error) {
	if req.Id <= 0 {
		return nil, errors.New("invalid task ID")
	}
	task, err := service.repo.GetTaskByID(ctx, int64(req.Id))
	if err != nil {
		return nil, fmt.Errorf("task with ID %d not found", req.Id)
	}
	return &v1.GetTaskByIdResponse{Task: task.ToProto()}, nil
}

func (service *taskService) DeleteTaskByID(ctx context.Context, req *v1.DeleteTaskByIdRequest) (*v1.DeleteTaskByIdResponse, error) {
	if req.Id <= 0 {
		return nil, errors.New("invalid task ID")
	}
	task, err := service.repo.GetTaskByID(ctx, int64(req.Id))

	if err != nil {
		return nil, fmt.Errorf("task with ID %d not found", task.ID)
	}

	err = service.repo.DeleteTaskByID(ctx, task.ID)
	if err != nil {
		fmt.Printf("error in DeleteID repo")
		return nil, err
	}
	return &v1.DeleteTaskByIdResponse{Message: "Task deleted successfully"}, nil
}
func (service *taskService) UpdateTaskByID(ctx context.Context, task *v1.UpdateTaskRequest) (*v1.UpdateTaskResponse, error) {
	if task.Task == nil {
		return nil, errors.New("task data is required") //avoid panic
	}
	taskModel := models.ProtoToModel(task.Task)
	err := service.repo.UpdateTaskByID(ctx, taskModel)
	if err != nil {
		fmt.Printf("Error Update in repo: %v\n", err)
		return nil, err

	}
	return &v1.UpdateTaskResponse{Task: taskModel.ToProto()}, nil
}
