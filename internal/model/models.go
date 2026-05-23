package model

import "time"

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleSupervisor Role = "supervisor"
	RoleEngineer  Role = "engineer"
)

type TicketStatus string

const (
	StatusCreated    TicketStatus = "created"
	StatusPending    TicketStatus = "pending"
	StatusAssigned   TicketStatus = "assigned"
	StatusProcessing TicketStatus = "processing"
	StatusSuspended  TicketStatus = "suspended"
	StatusReview     TicketStatus = "review"
	StatusCompleted  TicketStatus = "completed"
	StatusArchived   TicketStatus = "archived"
)

type TicketType string

const (
	TicketTypeFault   TicketType = "fault"
	TicketTypeChange  TicketType = "change"
	TicketTypeRequest TicketType = "request"
	TicketTypePatrol  TicketType = "patrol"
)

type Priority string

const (
	PriorityP0 Priority = "p0"
	PriorityP1 Priority = "p1"
	PriorityP2 Priority = "p2"
	PriorityP3 Priority = "p3"
)

type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"`
	RealName  string    `json:"real_name" db:"real_name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Role      Role      `json:"role" db:"role"`
	TeamID    *int64    `json:"team_id" db:"team_id"`
	Skills    []byte    `json:"skills" db:"skills"`
	Status    int       `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Team struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	SupervisorID *int64    `json:"supervisor_id" db:"supervisor_id"`
	Description  string    `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Ticket struct {
	ID          int64        `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Type        TicketType   `json:"type" db:"type"`
	Priority    Priority     `json:"priority" db:"priority"`
	Status      TicketStatus `json:"status" db:"status"`
	CreatorID   int64        `json:"creator_id" db:"creator_id"`
	AssigneeID  *int64       `json:"assignee_id" db:"assignee_id"`
	ProjectID   *int64       `json:"project_id" db:"project_id"`
	AssetID     *int64       `json:"asset_id" db:"asset_id"`
	AlertID     *int64       `json:"alert_id" db:"alert_id"`
	SLADeadline *time.Time   `json:"sla_deadline" db:"sla_deadline"`
	Metadata    []byte       `json:"metadata" db:"metadata"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

type TicketLog struct {
	ID        int64     `json:"id" db:"id"`
	TicketID  int64     `json:"ticket_id" db:"ticket_id"`
	Action    string    `json:"action" db:"action"`
	OperatorID int64   `json:"operator_id" db:"operator_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Project struct {
	ID              int64      `json:"id" db:"id"`
	Code            string     `json:"code" db:"code"`
	Name            string     `json:"name" db:"name"`
	Description     string     `json:"description" db:"description"`
	Type            string     `json:"type" db:"type"`
	Priority        string     `json:"priority" db:"priority"`
	Status          string     `json:"status" db:"status"`
	Requester       string     `json:"requester" db:"requester"`
	ManagerID       int64      `json:"manager_id" db:"manager_id"`
	Budget          *float64   `json:"budget" db:"budget"`
	Remark          string     `json:"remark" db:"remark"`
	StartDate       *time.Time `json:"start_date" db:"start_date"`
	EndDate         *time.Time `json:"end_date" db:"end_date"`
	ActualEndDate   *time.Time `json:"actual_end_date" db:"actual_end_date"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type ProjectMember struct {
	ID        int64     `json:"id" db:"id"`
	ProjectID int64     `json:"project_id" db:"project_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Score struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	TicketID  *int64    `json:"ticket_id" db:"ticket_id"`
	Period    string    `json:"period" db:"period"`
	Dimension string    `json:"dimension" db:"dimension"`
	Score     float64   `json:"score" db:"score"`
	Remark    string    `json:"remark" db:"remark"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type KnowledgeArticle struct {
	ID              int64     `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Content         string    `json:"content" db:"content"`
	Tags            []byte    `json:"tags" db:"tags"`
	AuthorID        int64     `json:"author_id" db:"author_id"`
	RelatedTickets  []byte    `json:"related_tickets" db:"related_tickets"`
	ViewCount       int       `json:"view_count" db:"view_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Schedule struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Date      string    `json:"date" db:"date"`
	ShiftType string    `json:"shift_type" db:"shift_type"`
	Remark    string    `json:"remark" db:"remark"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Asset struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	IP        string    `json:"ip" db:"ip"`
	Status    string    `json:"status" db:"status"`
	Location  string    `json:"location" db:"location"`
	Metadata  []byte    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MonitorAlert struct {
	ID          int64     `json:"id" db:"id"`
	Source      string    `json:"source" db:"source"`
	AlertName   string    `json:"alert_name" db:"alert_name"`
	Level       string    `json:"level" db:"level"`
	TicketID    *int64    `json:"ticket_id" db:"ticket_id"`
	RawPayload  []byte    `json:"raw_payload" db:"raw_payload"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Notification struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TicketCompletion struct {
	ID          int64     `json:"id" db:"id"`
	TicketID    int64     `json:"ticket_id" db:"ticket_id"`
	Solution    string    `json:"solution" db:"solution"`
	RootCause   string    `json:"root_cause" db:"root_cause"`
	Result      string    `json:"result" db:"result"`
	Impact      string    `json:"impact" db:"impact"`
	Remaining   string    `json:"remaining" db:"remaining"`
	Suggestion  string    `json:"suggestion" db:"suggestion"`
	Handover    string    `json:"handover" db:"handover"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TicketFile struct {
	ID         int64     `json:"id" db:"id"`
	TicketID   int64     `json:"ticket_id" db:"ticket_id"`
	Filename   string    `json:"filename" db:"filename"`
	Filepath   string    `json:"filepath" db:"filepath"`
	Filesize   int64     `json:"filesize" db:"filesize"`
	UploaderID int64     `json:"uploader_id" db:"uploader_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type OperationLog struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	ResourceID int64   `json:"resource_id" db:"resource_id"`
	Detail    string    `json:"detail" db:"detail"`
	IP        string    `json:"ip" db:"ip"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
