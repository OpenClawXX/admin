package repository

import (
	"ops-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	if user.Skills == nil {
		user.Skills = []byte("[]")
	}
	query := `INSERT INTO users (username, password, real_name, email, phone, role, team_id, skills, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		user.Username, user.Password, user.RealName, user.Email, user.Phone,
		user.Role, user.TeamID, user.Skills, user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(page, size int, teamID *int64) ([]model.User, int64, error) {
	var total int64
	var err error
	var users []model.User
	offset := (page - 1) * size

	if teamID != nil {
		err = r.db.Get(&total, "SELECT COUNT(*) FROM users WHERE team_id = $1", *teamID)
		if err != nil {
			return nil, 0, err
		}
		err = r.db.Select(&users,
			"SELECT * FROM users WHERE team_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3", *teamID, size, offset)
	} else {
		err = r.db.Get(&total, "SELECT COUNT(*) FROM users")
		if err != nil {
			return nil, 0, err
		}
		err = r.db.Select(&users,
			"SELECT * FROM users ORDER BY id DESC LIMIT $1 OFFSET $2", size, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) Update(user *model.User) error {
	if user.Skills == nil || len(user.Skills) == 0 {
		user.Skills = []byte("[]")
	}
	query := `UPDATE users SET real_name=$1, email=$2, phone=$3, role=$4, team_id=$5, skills=$6, status=$7, updated_at=NOW()
		WHERE id=$8`
	_, err := r.db.Exec(query,
		user.RealName, user.Email, user.Phone, user.Role, user.TeamID, user.Skills, user.Status, user.ID)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (r *UserRepository) ListByTeamID(teamID int64) ([]model.User, error) {
	var users []model.User
	err := r.db.Select(&users, "SELECT * FROM users WHERE team_id = $1 ORDER BY id", teamID)
	return users, err
}

func (r *UserRepository) UpdateTeamID(userID, teamID int64) error {
	_, err := r.db.Exec("UPDATE users SET team_id = $1, updated_at = NOW() WHERE id = $2", teamID, userID)
	return err
}

func (r *UserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	_, err := r.db.Exec("UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2", hashedPassword, userID)
	return err
}
