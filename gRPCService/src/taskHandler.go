package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/models"
	repo "jwt2/repo/src"
	"log"
	"strconv"
	"time"

	"github.com/bufbuild/protovalidate-go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskService interface {
	v1.TaskServiceServer
}
type TaskRepoInterface interface {
	CreateTask(ctx context.Context, task *models.Task) error

	GetAllTask(ctx context.Context, userID string, page *models.Paging) ([]models.Task, error)
	GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error)

	DeleteTaskByID(ctx context.Context, taskID int64) error
	UpdateTaskByID(ctx context.Context, task *models.Task) error
}
type TaskHandler struct {
	v1.UnimplementedTaskServiceServer //tránh lỗi khi không implement đủ các method
	repo                              TaskRepoInterface
	//RedisClient                       *redis.Client
	redisRepo repo.RedisClientInterface
}

func NewTaskHandler(repo TaskRepoInterface, redisRepo repo.RedisClientInterface) *TaskHandler {
	return &TaskHandler{repo: repo, redisRepo: redisRepo}
}

func (h *TaskHandler) PublishEvent(eventType string, taskID int64) error {
	str := strconv.FormatInt(taskID, 10)
	return h.redisRepo.Pub(context.Background(), "task_event", eventType+"|"+str)
}
func (h *TaskHandler) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskResponse, error) {
	//validate
	validator, err := protovalidate.New()
	if err != nil {
		return nil, errors.New("failed to initialize validator")
	}
	err = validator.Validate(req)
	if err != nil {
		log.Printf("validation err: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	//

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("some thing wrong with user id")
	}

	task := models.ProtoToModel(req.Task)
	task.UserID = userID
	err = h.repo.CreateTask(ctx, task)
	if err != nil {
		fmt.Printf("error creating task: %v", err)
		return nil, err
	}
	//
	str := strconv.FormatInt(task.ID, 10)
	key := "task:" + str
	jsonData, err := json.Marshal(task)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	err = h.redisRepo.Set(ctx, key, jsonData, time.Minute*30)
	if err != nil {
		log.Printf("error saving new task to redis: %v", err)
		return nil, err
	}
	//
	err = h.PublishEvent("task_created", task.ID)
	if err != nil {
		fmt.Printf("Failed to publish event: %v", err)
	}

	return &v1.CreateTaskResponse{Task: task.ToProto()}, nil
}

func (h *TaskHandler) GetTask(req *v1.GetTaskRequest, stream v1.TaskService_GetTaskServer) error {
	ctx := stream.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return errors.New("some thing wrong with user id")
	}
	if req.UserID == "" {
		req.UserID = userID
	}
	//redis
	// key := "task:" + userID
	// taskJson, err := h.RedisClient.Get(ctx, key).Result()
	// if err == nil {
	// 	var tasks []models.Task
	// 	json.Unmarshal([]byte(taskJson), &tasks)
	// 	var protoTask []*v1.Task
	// 	for _, task := range tasks {
	// 		protoTask = append(protoTask, task.ToProto())
	// 	}

	// 	return &v1.GetTaskResponse{Task: protoTask}, nil
	// }
	//

	page := 10
	limit := 1
	if req.Page != nil {
		if req.Page.Limit > 0 {
			page = int(req.Page.Page)
		}
		if req.Page.Limit > 0 {
			limit = int(req.Page.Limit)
		}
	}
	pageModel := &models.Paging{
		Page:      page,
		Limit:     limit,
		Title:     req.Page.Title,
		StartDate: req.Page.StartDate,
		EndDate:   req.Page.EndDate,
		Status:    req.Page.Status,
	}
	pageModel.Process()

	tasks, err := h.repo.GetAllTask(stream.Context(), req.UserID, pageModel)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get task: %v", err)
	}
	for _, task := range tasks {
		if err := stream.Send(&v1.GetTaskResponse{Task: task.ToProto()}); err != nil {
			return err
		}
	}

	//redis
	// var protoTask []*v1.Task
	// for _, task := range task {
	// 	protoTask = append(protoTask, task.ToProto())
	// }
	// toJson, err := json.Marshal(task)
	// if err == nil {

	// 	err = h.RedisClient.Set(ctx, key, toJson, time.Minute*30).Err()
	// 	if err != nil {
	// 		fmt.Printf("Failed to saving task to redis: %v", err)
	// 	}
	// }
	//
	return nil
}
func (h *TaskHandler) GetTaskById(ctx context.Context, req *v1.GetTaskByIdRequest) (*v1.GetTaskByIdResponse, error) {
	UserID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("some thing wrong with user_id in token")
	}
	if req.Id <= 0 {
		return nil, errors.New("invalid task id")
	}
	//Redis
	key := "task:" + UserID
	taskJson, err := h.redisRepo.Get(ctx, key)
	if err != nil {
		fmt.Printf("task data in redis is empty\n")
	}
	var tasks []models.Task
	json.Unmarshal([]byte(taskJson), &tasks)
	for _, task := range tasks {
		if task.ID == int64(req.Id) {

			return &v1.GetTaskByIdResponse{Task: task.ToProto()}, nil
		}
	}

	//
	task, err := h.repo.GetTaskByID(ctx, int64(req.Id))

	if err != nil {
		return nil, fmt.Errorf("task with ID %d not found", req.Id)
	}
	if task.UserID != UserID {
		return nil, errors.New("Forbidden")
	}
	jsonData, err := json.Marshal(task)
	if err != nil {
		return nil, errors.New("Failed to marshal data redis: " + err.Error())
	}
	err = h.redisRepo.Set(ctx, key, jsonData, time.Minute*30)
	if err != nil {
		log.Printf("you cooked: %v", err)
		return nil, err
	}
	return &v1.GetTaskByIdResponse{Task: task.ToProto()}, nil
}

func (h *TaskHandler) DeleteTaskByID(ctx context.Context, req *v1.DeleteTaskByIdRequest) (*v1.DeleteTaskByIdResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("some thing wrong with user id in token")
	}
	if req.Id <= 0 {
		return nil, errors.New("invalid task ID")
	}
	task, err := h.repo.GetTaskByID(ctx, int64(req.Id))
	if err != nil {
		return nil, fmt.Errorf("task with ID %d not found", task.ID)
	}
	if task.UserID != userId {
		return nil, errors.New("Forbidden")
	}
	err = h.repo.DeleteTaskByID(ctx, task.ID)
	if err != nil {
		return nil, errors.New("error in delete task by ID: %s" + err.Error())
	}
	//redis
	err = h.PublishEvent("task_deleted", int64(req.Id))
	if err != nil {
		fmt.Printf("Failed to publish event: %v", err)
	}

	key := "task:" + userId
	err = h.redisRepo.Del(ctx, key)
	if err != nil {
		fmt.Printf("some thing wrong when delete data in redis")
	}
	//
	return &v1.DeleteTaskByIdResponse{Message: "Task deleted successfully"}, nil
}
func (h *TaskHandler) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.UpdateTaskResponse, error) {
	userId, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("some thing wrong with user id in token")
	}
	if req.Task == nil {
		return nil, errors.New("task data is required")
	}
	taskB, err := h.repo.GetTaskByID(ctx, int64(req.Task.Id))
	if err != nil {
		return nil, errors.New("some thing wrong in get task " + err.Error())
	}
	if taskB.UserID != userId {
		return nil, errors.New("Forbidden aa" + taskB.UserID)
	}

	taskModel := models.ProtoToModel(req.Task)
	taskModel.UserID = userId
	if req.Task.CreatedAt == "" {
		taskModel.CreatedAt = taskB.CreatedAt
	}

	err = h.repo.UpdateTaskByID(ctx, taskModel)
	if err != nil {
		return nil, err
	}
	//redis
	key := "task:" + userId
	err = h.redisRepo.Del(ctx, key)
	if err != nil {
		fmt.Printf("some thing wrong when update data in redis")
	}
	err = h.PublishEvent("task_updated", int64(req.Task.Id))
	if err != nil {
		fmt.Printf("Failed to publish event: %v", err)
	}
	//
	return &v1.UpdateTaskResponse{Task: taskModel.ToProto()}, nil
}
