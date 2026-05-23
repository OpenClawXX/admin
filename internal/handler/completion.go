package handler

import (
	"net/http"
	"strconv"

	"ops-platform/internal/model"
	"ops-platform/internal/pkg/response"
	"ops-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type CompletionHandler struct {
	completionService *service.CompletionService
}

func NewCompletionHandler(completionService *service.CompletionService) *CompletionHandler {
	return &CompletionHandler{completionService: completionService}
}

func (h *CompletionHandler) Submit(c *gin.Context) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	var req struct {
		Solution   string `json:"solution" binding:"required"`
		RootCause  string `json:"root_cause"`
		Result     string `json:"result"`
		Impact     string `json:"impact"`
		Remaining  string `json:"remaining"`
		Suggestion string `json:"suggestion"`
		Handover   string `json:"handover"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写解决方案")
		return
	}

	completion := &model.TicketCompletion{
		TicketID:   ticketID,
		Solution:   req.Solution,
		RootCause:  req.RootCause,
		Result:     req.Result,
		Impact:     req.Impact,
		Remaining:  req.Remaining,
		Suggestion: req.Suggestion,
		Handover:   req.Handover,
	}

	if err := h.completionService.SaveCompletion(completion); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, completion)
}

func (h *CompletionHandler) Get(c *gin.Context) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	completion, err := h.completionService.GetCompletion(ticketID)
	if err != nil {
		response.Success(c, nil)
		return
	}

	response.Success(c, completion)
}

func (h *CompletionHandler) UploadFile(c *gin.Context) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	uploaderID := c.GetInt64("user_id")

	f, err := h.completionService.SaveFile(ticketID, uploaderID, header.Filename, file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, f)
}

func (h *CompletionHandler) ListFiles(c *gin.Context) {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的工单ID")
		return
	}

	files, err := h.completionService.GetFiles(ticketID)
	if err != nil {
		response.Success(c, []interface{}{})
		return
	}
	if files == nil {
		files = []model.TicketFile{}
	}

	response.Success(c, files)
}

func (h *CompletionHandler) DownloadFile(c *gin.Context) {
	fileID, err := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	f, err := h.completionService.GetFile(fileID)
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+f.Filename)
	c.File(f.Filepath)
}

func (h *CompletionHandler) DeleteFile(c *gin.Context) {
	fileID, err := strconv.ParseInt(c.Param("file_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}

	f, err := h.completionService.GetFile(fileID)
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}

	userID := c.GetInt64("user_id")
	role, _ := c.Get("role")
	userRole := role.(model.Role)

	if f.UploaderID != userID && userRole != model.RoleAdmin && userRole != model.RoleSupervisor {
		response.Forbidden(c, "只能删除自己上传的文件")
		return
	}

	if err := h.completionService.DeleteFile(fileID); err != nil {
		response.InternalError(c, "删除失败")
		return
	}

	response.Success(c, nil)
}

func (h *CompletionHandler) UploadPage(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}
