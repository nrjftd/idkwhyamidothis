package repo

import (
	"context"
	"database/sql"
	"jwt2/models"
)

// Interface cho Select Query
type SelectQuerier interface {
	Model(model interface{}) SelectQuerier
	Where(query string, args ...interface{}) SelectQuerier
	Order(order string) SelectQuerier
	Limit(limit int) SelectQuerier
	Offset(offSet int) SelectQuerier
	Scan(ctx context.Context, dest interface{}) error
}

// Interface cho Insert Query
type InsertQuerier interface {
	Model(model interface{}) InsertQuerier
	Exec(ctx context.Context) (sql.Result, error)
}
type DeleteQuerier interface {
	Model(model interface{}) DeleteQuerier
	Where(query string, args ...interface{}) DeleteQuerier
	Exec(ctx context.Context) (sql.Result, error)
}
type UpdateQuerier interface {
	Model(model interface{}) UpdateQuerier
	Where(query string, args ...interface{}) UpdateQuerier
	Exec(ctx context.Context) (sql.Result, error)
}

// Interface cho DB operations
type BunDB interface {
	NewSelect() SelectQuerier
	NewInsert() InsertQuerier
	NewDelete() DeleteQuerier
	NewUpdate() UpdateQuerier
}

// Interface định nghĩa các operations của Task
type TaskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetAllTask(ctx context.Context, userID string, page *models.Paging) ([]models.Task, error)
	GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error)
	DeleteTaskByID(ctx context.Context, taskID int64) error
	UpdateTaskByID(ctx context.Context, task *models.Task) error
}

type TaskRepo struct {
	db BunDB
}

func NewTaskRepo(db BunDB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (repo *TaskRepo) CreateTask(ctx context.Context, task *models.Task) error {
	_, err := repo.db.NewInsert().Model(task).Exec(ctx)
	return err
}

func (repo *TaskRepo) GetAllTask(ctx context.Context, userID string, page *models.Paging) ([]models.Task, error) {
	var tasks []models.Task
	page.Process()
	query := repo.db.NewSelect().Model(&tasks).Where("user_id =?", userID)
	if page.Title != "" {
		query.Where("title ILike ?", "%"+page.Title+"%")
	}
	if page.StartDate != "" && page.EndDate != "" {
		query.Where("created_at BETWEEN ? AND ?", page.StartDate, page.EndDate)
	}
	if page.StartDate != "" {
		query.Where("created_at >=?", page.StartDate)
	}
	if page.EndDate != "" {
		query.Where("created_at <=?", page.EndDate)
	}
	err := query.Order("id ASC").Limit(page.Limit).Offset(page.OffSet()).Scan(ctx, &tasks) //write order, where, offset, limit in wrapper

	return tasks, err
}

func (repo *TaskRepo) GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error) {
	var task models.Task
	err := repo.db.NewSelect().Model(&task).Where("id =?", taskID).Scan(ctx, &task)
	return &task, err
}

func (repo *TaskRepo) DeleteTaskByID(ctx context.Context, TaskID int64) error {
	_, err := repo.db.NewDelete().Model((*models.Task)(nil)).Where("id =?", TaskID).Exec(ctx)
	return err
}

func (repo *TaskRepo) UpdateTaskByID(ctx context.Context, task *models.Task) error {
	_, err := repo.db.NewUpdate().Model(task).Where("id =?", task.ID).Exec(ctx)
	return err

}
