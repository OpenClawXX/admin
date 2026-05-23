package handler

import (
	"strconv"

	"ops-platform/internal/model"
	"ops-platform/internal/pkg/response"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) Create(c *gin.Context) {
	var team model.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.teamService.Create(&team); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, team)
}

func (h *TeamHandler) List(c *gin.Context) {
	teams, err := h.teamService.List()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, teams)
}

func (h *TeamHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	team, err := h.teamService.GetByID(id)
	if err != nil {
		response.NotFound(c, "团队不存在")
		return
	}
	response.Success(c, team)
}

func (h *TeamHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var team model.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	team.ID = id

	if err := h.teamService.Update(&team); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.teamService.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}
