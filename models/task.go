package models

import (
	"encoding/json"
	"fmt"
	v1 "jwt2/gen/go/proto/task/v1"
	"time"

	"github.com/uptrace/bun"
)

type TaskStatusWrapper int32

func (ts *TaskStatusWrapper) UnmarshalJSON(data []byte) error {
	var intValue int32
	if err := json.Unmarshal(data, &intValue); err == nil {
		*ts = TaskStatusWrapper(intValue)
		return nil
	}
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		status, exists := stringToStatus[strValue]
		if !exists {
			return fmt.Errorf("invalid status value: %s. Please input 1 to 4 or 'Doing', 'Pending', 'Done', 'Cancelled'", strValue)
		}
		*ts = TaskStatusWrapper(status)
		return nil
	}
	return fmt.Errorf("invalid status format")
}

func (ts *TaskStatusWrapper) MarshalJSON() ([]byte, error) {
	str, exists := statusToString[v1.TaskStatus(*ts)]
	if !exists {
		return nil, fmt.Errorf("invalid status value: %d", ts)

	}
	return json.Marshal(str)
}

var statusToString = map[v1.TaskStatus]string{
	v1.TaskStatus_TASK_STATUS_PENDING:   "Pending",
	v1.TaskStatus_TASK_STATUS_DOING:     "Doing",
	v1.TaskStatus_TASK_STATUS_DONE:      "Done",
	v1.TaskStatus_TASK_STATUS_CANCELLED: "Cancelled",
}
var stringToStatus = map[string]v1.TaskStatus{
	"Pending":   v1.TaskStatus_TASK_STATUS_PENDING,
	"Doing":     v1.TaskStatus_TASK_STATUS_DOING,
	"Done":      v1.TaskStatus_TASK_STATUS_DONE,
	"Cancelled": v1.TaskStatus_TASK_STATUS_CANCELLED,
}

type Task struct {
	bun.BaseModel `bun:"table:task" swaggerignore:"true"`
	ID            int64             `bun:",pk,autoincrement" json:"id"`
	Title         string            `bun:"title,notnull" json:"title" validate:"required, min=2,max=100"`
	Description   string            `bun:"description,notnull" json:"description" validate:"required,min=10"`
	Status        TaskStatusWrapper `bun:"status,notnull" json:"status"`
	CreatedAt     time.Time         `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time         `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	UserID        string            `bun:"user_id,notnull" json:"user_id"`
}

// ===================================================================================================//

// ===================================================================================================//
func (t *Task) ToProto() *v1.Task {
	fmt.Println("Task status before conversion:", t.Status)

	return &v1.Task{
		Id:          int32(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Status:      v1.TaskStatus(t.Status),
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
		UserID:      t.UserID,
	}
}
func ProtoToModel(proto *v1.Task) *Task {
	var createAt, updateAt time.Time
	var err error
	if proto.CreatedAt != "" {
		createAt, err = time.Parse(time.RFC3339, proto.CreatedAt)
		if err != nil {
			fmt.Println("error parsing CreatedAt", err)
			createAt = time.Now()
		}
	} else {
		createAt = time.Now()
	}
	if proto.UpdatedAt != "" {

		updateAt, err = time.Parse(time.RFC3339, proto.UpdatedAt)
		if err != nil {
			fmt.Println("erorr parsing UpdateAt", err)
			updateAt = time.Now()
		}
	} else {
		updateAt = time.Now()
	}
	fmt.Println("Task Status received from Proto:", proto.Status)

	return &Task{
		ID:          int64(proto.Id),
		Title:       proto.Title,
		Description: proto.Description,
		Status:      TaskStatusWrapper(proto.GetStatus()),
		CreatedAt:   createAt,
		UpdatedAt:   updateAt,
		UserID:      proto.UserID,
	}
}
func IsValidStatus(status string) bool {
	validStatuses := []string{"Pending", "Doing", "Done", "Cancelled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}
