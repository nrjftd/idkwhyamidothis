package controller

import (
	"context"
	v1 "jwt2/gen/go/proto/task/v1"
	"jwt2/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskServiceInterface interface {
	CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskResponse, error)
	GetAllTask(ctx context.Context, userID string) (*v1.GetTaskResponse, error)
	GetTaskByID(ctx context.Context, req *v1.GetTaskByIdRequest) (*v1.GetTaskByIdResponse, error)
	DeleteTaskByID(ctx context.Context, req *v1.DeleteTaskByIdRequest) (*v1.DeleteTaskByIdResponse, error)
	UpdateTaskByID(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.UpdateTaskResponse, error)
}

type TaskController struct {
	//service TaskServiceInterface
	client v1.TaskServiceClient
}

// func NewTaskController(service TaskServiceInterface) *TaskController {
func NewTaskController(client v1.TaskServiceClient) *TaskController {
	//	return &TaskController{service: service}
	return &TaskController{client: client}
}

///===========================================================================================================================================================///

// CreateTask godoc
// @Summary Create new task
// @Description Create a new task for user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body models.Task true "Task data"
// @Success 201 {object} map[string]string "Task created successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /users/tasks [post]
func (control *TaskController) CreateTask(c *gin.Context) {

	var task models.Task
	err := c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return //thừa
	}
	userIDstr, ok := userID.(string)
	if !ok || userIDstr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse userID",
		})
		return
	}
	task.UserID = userIDstr
	//fmt.Println("Task Status:", task.Status)

	req := &v1.CreateTaskRequest{Task: task.ToProto()}
	//fmt.Println("Status received:", task.Status)

	resp, err := control.client.CreateTask(context.Background(), req) //
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task lmao" + err.Error(),
		})

		return
	}
	c.JSON(http.StatusCreated, resp.Task)

}

///===========================================================================================================================================================///

// GetAllTask godoc
// @Summary Get all task
// @Description Fetch a list of all tasks assigned to the authenticated user
// @Tags Tasks
// @Accept json
// @Produce json
// @Success 200 {array} models.Task "list of tasks"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /users/tasks [get]
// func (control *TaskController) GetAllTask(c *gin.Context) {
// 	userID, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "Unauthorized",
// 		})
// 		return
// 	}
// 	req := &v1.GetTaskRequest{UserID: userID.(string)}
// 	resp, err := control.client.GetTask(c, req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "Failed to fetch tasks",
// 		})
// 		return
// 	}
// 	tasks := make([]models.Task, len(resp.Task))
// 	for i, task := range resp.Task {
// 		tasks[i] = *models.ProtoToModel(task)
// 	}
// 	c.JSON(http.StatusOK, tasks)
// }

// GetTaskByID godoc
// @Summary Get Task By ID
// @Description Fetch task detail by Task ID for the authenticated user
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "task ID"
// @Success 200 {object} models.Task "Task details"
// @Failure 400 {object} map[string]string "Invalid Task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Security BearerAuth
// @Router /users/tasks/{id} [get]
func (control *TaskController) GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")
	id, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task ID",
		})
		return
	}
	req := &v1.GetTaskByIdRequest{Id: int32(id)}
	resp, err := control.client.GetTaskById(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, resp.Task)
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete a task by id
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "task ID"
// @Success 200 {object} map[string]string "Task deleted successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Sercurity BearerAuth
// @Router /users/tasks/delete/{id} [delete]
func (control *TaskController) DeleteTaskByID(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	//
	taskID := c.Param("id")
	id, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	getTask := &v1.GetTaskByIdRequest{Id: int32(id)}
	respGet, err := control.client.GetTaskById(context.Background(), getTask)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}
	if respGet.Task.UserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "You are not authorized to delete this task. Cook",
		})
		return
	}

	req := &v1.DeleteTaskByIdRequest{Id: int32(id)}
	resp, err := control.client.DeleteTaskByID(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateTaskByID godoc
// @Summary Update Task By ID
// @Description Delete task by ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Success 200 {object} models.Task "Updated successfully"
// @Failure 400 {object} map[string]string "Invalid Task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Sercurity BearerAuth
// @Router /users/tasks/update/{id}[put]

func (control *TaskController) UpdateTaskByID(c *gin.Context) {
	var task models.Task
	err := c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	taskID := c.Param("id")
	id, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	getTask := &v1.GetTaskByIdRequest{Id: int32(id)}
	respGet, err := control.client.GetTaskById(context.Background(), getTask)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}
	if respGet.Task.UserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "You are not authorized to update this task",
		})
		return
	}
	req := &v1.UpdateTaskRequest{Task: task.ToProto()}
	respU, errorr := control.client.UpdateTask(context.Background(), req)
	if errorr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update task: " + errorr.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, respU.Task)
}
