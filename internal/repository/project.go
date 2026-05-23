package repository

import (
	"fmt"

	"ops-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *model.Project) error {
	query := `INSERT INTO projects (code, name, description, type, priority, status, requester, manager_id, budget, remark, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		project.Code, project.Name, project.Description, project.Type, project.Priority,
		project.Status, project.Requester, project.ManagerID, project.Budget, project.Remark,
		project.StartDate, project.EndDate,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *ProjectRepository) FindByID(id int64) (*model.Project, error) {
	var project model.Project
	err := r.db.Get(&project, "SELECT * FROM projects WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) List() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Select(&projects, "SELECT * FROM projects ORDER BY id DESC")
	return projects, err
}

func (r *ProjectRepository) Update(project *model.Project) error {
	query := `UPDATE projects SET name=$1, description=$2, type=$3, priority=$4, status=$5, requester=$6,
		manager_id=$7, budget=$8, remark=$9, start_date=$10, end_date=$11, actual_end_date=$12, updated_at=NOW() WHERE id=$13`
	_, err := r.db.Exec(query,
		project.Name, project.Description, project.Type, project.Priority, project.Status,
		project.Requester, project.ManagerID, project.Budget, project.Remark,
		project.StartDate, project.EndDate, project.ActualEndDate, project.ID)
	return err
}

func (r *ProjectRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM project_members WHERE project_id = $1", id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM projects WHERE id = $1", id)
	return err
}

func (r *ProjectRepository) NextCode() (string, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM projects")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PRJ-%04d", count+1), nil
}

// Project members
func (r *ProjectRepository) SetMembers(projectID int64, userIDs []int64) error {
	_, err := r.db.Exec("DELETE FROM project_members WHERE project_id = $1", projectID)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		_, err := r.db.Exec("INSERT INTO project_members (project_id, user_id) VALUES ($1, $2)", projectID, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectRepository) GetMemberIDs(projectID int64) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids, "SELECT user_id FROM project_members WHERE project_id = $1", projectID)
	return ids, err
}
