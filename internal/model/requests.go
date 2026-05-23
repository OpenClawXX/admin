package model

import "encoding/json"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateTicketRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Type        TicketType `json:"type" binding:"required"`
	Priority    Priority   `json:"priority" binding:"required"`
	AssigneeID  *int64     `json:"assignee_id"`
	ProjectID   *int64     `json:"project_id"`
	AssetID     *int64     `json:"asset_id"`
}

type UpdateTicketRequest struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Type        *TicketType   `json:"type"`
	Priority    *Priority     `json:"priority"`
	Status      *TicketStatus `json:"status"`
	AssigneeID  *int64        `json:"assignee_id"`
	ProjectID   *int64        `json:"project_id"`
}

type AssignTicketRequest struct {
	AssigneeID int64  `json:"assignee_id" binding:"required"`
	Remark     string `json:"remark"`
}

type TransferTicketRequest struct {
	AssigneeID int64  `json:"assignee_id" binding:"required"`
	Remark     string `json:"remark"`
}

type AddProgressRequest struct {
	Content string `json:"content" binding:"required"`
	Status  string `json:"status"`
}

type AddLogRequest struct {
	Content string `json:"content" binding:"required"`
}

type CompleteTicketRequest struct {
	Solution string `json:"solution" binding:"required"`
	Remark   string `json:"remark"`
}

type TicketQuery struct {
	Status     string `form:"status"`
	Priority   string `form:"priority"`
	Type       string `form:"type"`
	AssigneeID *int64 `form:"assignee_id"`
	ProjectID  *int64 `form:"project_id"`
	Keyword    string `form:"keyword"`
	TeamID     *int64 `form:"-"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type CreateProjectRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	Priority      string   `json:"priority"`
	Requester     string   `json:"requester"`
	ManagerID     int64    `json:"manager_id" binding:"required"`
	MemberIDs     []int64  `json:"member_ids"`
	Budget        *float64 `json:"budget"`
	Remark        string   `json:"remark"`
	StartDate     *string  `json:"start_date"`
	EndDate       *string  `json:"end_date"`
}

type CreateArticleRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

type CreateAssetRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	IP       string `json:"ip"`
	Location string `json:"location"`
}

type WebhookAlertRequest struct {
	Source    string          `json:"source" binding:"required"`
	AlertName string          `json:"alert_name" binding:"required"`
	Level     string          `json:"level" binding:"required"`
	Payload   json.RawMessage `json:"payload"`
}
