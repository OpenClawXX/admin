package repository

import (
	"fmt"
	"strings"

	"ops-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type TicketRepository struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ticket *model.Ticket) error {
	if ticket.Metadata == nil {
		ticket.Metadata = []byte("{}")
	}
	query := `INSERT INTO tickets (title, description, type, priority, status, creator_id, assignee_id, project_id, asset_id, alert_id, sla_deadline, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		ticket.Title, ticket.Description, ticket.Type, ticket.Priority, ticket.Status,
		ticket.CreatorID, ticket.AssigneeID, ticket.ProjectID, ticket.AssetID, ticket.AlertID,
		ticket.SLADeadline, ticket.Metadata,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
}

func (r *TicketRepository) FindByID(id int64) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.Get(&ticket, "SELECT * FROM tickets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) List(query *model.TicketQuery, userID int64, role model.Role) ([]model.Ticket, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if query.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, query.Status)
		argIdx++
	}
	if query.Priority != "" {
		where = append(where, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, query.Priority)
		argIdx++
	}
	if query.Type != "" {
		where = append(where, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, query.Type)
		argIdx++
	}
	if query.AssigneeID != nil {
		where = append(where, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *query.AssigneeID)
		argIdx++
	}
	if query.ProjectID != nil {
		where = append(where, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, *query.ProjectID)
		argIdx++
	}
	if query.Keyword != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+query.Keyword+"%")
		argIdx++
	}

	// Data permission
	if role == model.RoleEngineer {
		where = append(where, fmt.Sprintf("(tickets.creator_id = $%d OR tickets.assignee_id = $%d)", argIdx, argIdx))
		args = append(args, userID)
		argIdx++
	} else if role == model.RoleSupervisor && query.TeamID != nil {
		where = append(where, fmt.Sprintf(`(tickets.creator_id IN (SELECT id FROM users WHERE team_id = $%d)
			OR tickets.assignee_id IN (SELECT id FROM users WHERE team_id = $%d))`, argIdx, argIdx))
		args = append(args, *query.TeamID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets WHERE %s", whereClause)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	args = append(args, query.PageSize, offset)
	listQuery := fmt.Sprintf("SELECT * FROM tickets WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, argIdx, argIdx+1)

	var tickets []model.Ticket
	if err := r.db.Select(&tickets, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *TicketRepository) UpdateStatus(id int64, status model.TicketStatus) error {
	_, err := r.db.Exec("UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

func (r *TicketRepository) UpdateAssignee(id int64, assigneeID int64) error {
	_, err := r.db.Exec("UPDATE tickets SET assignee_id = $1, status = $2, updated_at = NOW() WHERE id = $3",
		assigneeID, model.StatusAssigned, id)
	return err
}

func (r *TicketRepository) Update(ticket *model.Ticket) error {
	query := `UPDATE tickets SET title=$1, description=$2, type=$3, priority=$4, status=$5,
		assignee_id=$6, project_id=$7, updated_at=NOW() WHERE id=$8`
	_, err := r.db.Exec(query,
		ticket.Title, ticket.Description, ticket.Type, ticket.Priority, ticket.Status,
		ticket.AssigneeID, ticket.ProjectID, ticket.ID)
	return err
}

// TicketLog operations
func (r *TicketRepository) AddLog(log *model.TicketLog) error {
	query := `INSERT INTO ticket_logs (ticket_id, action, operator_id, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, log.TicketID, log.Action, log.OperatorID, log.Content).
		Scan(&log.ID, &log.CreatedAt)
}

func (r *TicketRepository) GetLogs(ticketID int64) ([]model.TicketLog, error) {
	var logs []model.TicketLog
	err := r.db.Select(&logs, "SELECT * FROM ticket_logs WHERE ticket_id = $1 ORDER BY created_at DESC", ticketID)
	return logs, err
}

func (r *TicketRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM ticket_logs WHERE ticket_id = $1", id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM tickets WHERE id = $1", id)
	return err
}
