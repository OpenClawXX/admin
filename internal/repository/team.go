package repository

import (
	"ops-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *model.Team) error {
	query := `INSERT INTO teams (name, supervisor_id, description)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, team.Name, team.SupervisorID, team.Description).
		Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *TeamRepository) FindByID(id int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(&team, "SELECT * FROM teams WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) List() ([]model.Team, error) {
	var teams []model.Team
	err := r.db.Select(&teams, "SELECT * FROM teams ORDER BY id")
	return teams, err
}

func (r *TeamRepository) Update(team *model.Team) error {
	query := `UPDATE teams SET name=$1, supervisor_id=$2, description=$3, updated_at=NOW() WHERE id=$4`
	_, err := r.db.Exec(query, team.Name, team.SupervisorID, team.Description, team.ID)
	return err
}

func (r *TeamRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM teams WHERE id = $1", id)
	return err
}

func (r *TeamRepository) FindBySupervisorID(supervisorID int64) (*model.Team, error) {
	var team model.Team
	err := r.db.Get(&team, "SELECT * FROM teams WHERE supervisor_id = $1", supervisorID)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) ClearSupervisor(teamID int64) error {
	_, err := r.db.Exec("UPDATE teams SET supervisor_id = NULL, updated_at = NOW() WHERE id = $1", teamID)
	return err
}
