package handler

import (
	"strconv"

	"ops-platform/internal/model"
	"ops-platform/internal/pkg/response"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *UserHandler) Create(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.userService.Create(&user); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var teamID *int64
	if tid := c.Query("team_id"); tid != "" {
		if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
			teamID = &id
		}
	}

	users, total, err := h.userService.List(page, size, teamID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, model.NewPageResult(users, total, page, size))
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	user.ID = id

	if err := h.userService.Update(&user); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := h.userService.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	plainPassword, err := h.userService.ResetPassword(id, req.Password)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"password": plainPassword})
}
