package handler

import (
	"strconv"

	"ops-platform/internal/model"
	"ops-platform/internal/pkg/response"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketService *service.TicketService
	userService   *service.UserService
}

func NewTicketHandler(ticketService *service.TicketService, userService *service.UserService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService, userService: userService}
}

func (h *TicketHandler) Create(c *gin.Context) {
	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	ticket := &model.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		ProjectID:   req.ProjectID,
		AssetID:     req.AssetID,
	}

	if err := h.ticketService.Create(ticket, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ticket)
}

func (h *TicketHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	ticket, err := h.ticketService.GetByID(id)
	if err != nil {
		response.NotFound(c, "工单不存在")
		return
	}

	response.Success(c, ticket)
}

func (h *TicketHandler) List(c *gin.Context) {
	var query model.TicketQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查询参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	role, _ := c.Get("role")
	userRole := role.(model.Role)

	if userRole == model.RoleSupervisor {
		if user, err := h.userService.GetByID(userID); err == nil && user.TeamID != nil {
			query.TeamID = user.TeamID
		}
	}

	tickets, total, err := h.ticketService.List(&query, userID, userRole)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, model.NewPageResult(tickets, total, query.Page, query.PageSize))
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	ticket, err := h.ticketService.GetByID(id)
	if err != nil {
		response.NotFound(c, "工单不存在")
		return
	}

	if req.Title != nil {
		ticket.Title = *req.Title
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Type != nil {
		ticket.Type = *req.Type
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		ticket.AssigneeID = req.AssigneeID
	}
	if req.ProjectID != nil {
		ticket.ProjectID = req.ProjectID
	}

	if err := h.ticketService.Update(ticket); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Assign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Assign(id, req.AssigneeID, operatorID, req.Remark); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Transfer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.TransferTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Transfer(id, req.AssigneeID, operatorID, req.Remark); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Suspend(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Suspend(id, operatorID, req.Reason); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Resume(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Resume(id, operatorID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) AddProgress(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.AddProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.AddProgress(id, operatorID, req.Content, req.Status); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) AddLog(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.AddLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.AddLog(id, operatorID, req.Content); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) GetLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	logs, err := h.ticketService.GetLogs(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, logs)
}

func (h *TicketHandler) Complete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.CompleteTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Complete(id, operatorID, req.Solution, req.Remark); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Review(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Review(id, operatorID, req.Approved, req.Remark); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Archive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	operatorID := c.GetInt64("user_id")
	if err := h.ticketService.Archive(id, operatorID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *TicketHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.ticketService.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
