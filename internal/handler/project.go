package handler

import (
	"strconv"

	"ops-platform/internal/model"
	"ops-platform/internal/pkg/response"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	project := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Priority:    req.Priority,
		Requester:   req.Requester,
		ManagerID:   req.ManagerID,
		Budget:      req.Budget,
		Remark:      req.Remark,
	}

	if err := h.projectService.Create(project, req.MemberIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, project)
}

func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.projectService.List()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, projects)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	project, err := h.projectService.GetByID(id)
	if err != nil {
		response.NotFound(c, "项目不存在")
		return
	}
	memberIDs, _ := h.projectService.GetMemberIDs(id)
	result := gin.H{
		"project":     project,
		"member_ids":  memberIDs,
	}
	response.Success(c, result)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req model.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	project := &model.Project{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Priority:    req.Priority,
		Requester:   req.Requester,
		ManagerID:   req.ManagerID,
		Budget:      req.Budget,
		Remark:      req.Remark,
	}

	if err := h.projectService.Update(project, req.MemberIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := h.projectService.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}
