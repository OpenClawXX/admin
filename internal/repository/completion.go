package repository

import (
	"ops-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type CompletionRepository struct {
	db *sqlx.DB
}

func NewCompletionRepository(db *sqlx.DB) *CompletionRepository {
	return &CompletionRepository{db: db}
}

func (r *CompletionRepository) Create(c *model.TicketCompletion) error {
	query := `INSERT INTO ticket_completions (ticket_id, solution, root_cause, result, impact, remaining, suggestion, handover)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (ticket_id) DO UPDATE SET
			solution = EXCLUDED.solution,
			root_cause = EXCLUDED.root_cause,
			result = EXCLUDED.result,
			impact = EXCLUDED.impact,
			remaining = EXCLUDED.remaining,
			suggestion = EXCLUDED.suggestion,
			handover = EXCLUDED.handover,
			created_at = NOW()
		RETURNING id, created_at`
	return r.db.QueryRow(query, c.TicketID, c.Solution, c.RootCause, c.Result,
		c.Impact, c.Remaining, c.Suggestion, c.Handover,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *CompletionRepository) FindByTicketID(ticketID int64) (*model.TicketCompletion, error) {
	var c model.TicketCompletion
	err := r.db.Get(&c, "SELECT * FROM ticket_completions WHERE ticket_id = $1", ticketID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// File operations
func (r *CompletionRepository) CreateFile(f *model.TicketFile) error {
	query := `INSERT INTO ticket_files (ticket_id, filename, filepath, filesize, uploader_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRow(query, f.TicketID, f.Filename, f.Filepath, f.Filesize, f.UploaderID,
	).Scan(&f.ID, &f.CreatedAt)
}

func (r *CompletionRepository) FindFilesByTicketID(ticketID int64) ([]model.TicketFile, error) {
	var files []model.TicketFile
	err := r.db.Select(&files, "SELECT * FROM ticket_files WHERE ticket_id = $1 ORDER BY created_at", ticketID)
	return files, err
}

func (r *CompletionRepository) FindFileByID(id int64) (*model.TicketFile, error) {
	var f model.TicketFile
	err := r.db.Get(&f, "SELECT * FROM ticket_files WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *CompletionRepository) DeleteFile(id int64) error {
	_, err := r.db.Exec("DELETE FROM ticket_files WHERE id = $1", id)
	return err
}
