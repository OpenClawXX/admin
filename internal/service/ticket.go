package service

import (
	"errors"

	"ops-platform/internal/model"
	"ops-platform/internal/repository"
)

type TicketService struct {
	ticketRepo     *repository.TicketRepository
	completionSvc  *CompletionService
}

func NewTicketService(ticketRepo *repository.TicketRepository) *TicketService {
	return &TicketService{ticketRepo: ticketRepo}
}

func (s *TicketService) SetCompletionService(svc *CompletionService) {
	s.completionSvc = svc
}

func (s *TicketService) Create(ticket *model.Ticket, creatorID int64) error {
	ticket.CreatorID = creatorID
	ticket.Status = model.StatusCreated

	if ticket.AssigneeID != nil {
		ticket.Status = model.StatusAssigned
	}

	if err := s.ticketRepo.Create(ticket); err != nil {
		return err
	}

	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticket.ID,
		Action:     "created",
		OperatorID: creatorID,
		Content:    "工单已创建",
	})
}

func (s *TicketService) GetByID(id int64) (*model.Ticket, error) {
	return s.ticketRepo.FindByID(id)
}

func (s *TicketService) List(query *model.TicketQuery, userID int64, role model.Role) ([]model.Ticket, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	return s.ticketRepo.List(query, userID, role)
}

func (s *TicketService) Assign(ticketID, assigneeID, operatorID int64, remark string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status == model.StatusCompleted || ticket.Status == model.StatusArchived {
		return errors.New("已完单或归档的工单不能派发")
	}

	if err := s.ticketRepo.UpdateAssignee(ticketID, assigneeID); err != nil {
		return err
	}

	content := "工单已派发"
	if remark != "" {
		content += "：" + remark
	}
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "assigned",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Transfer(ticketID, newAssigneeID, operatorID int64, remark string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status == model.StatusCompleted || ticket.Status == model.StatusArchived {
		return errors.New("已完单或归档的工单不能转派")
	}

	if err := s.ticketRepo.UpdateAssignee(ticketID, newAssigneeID); err != nil {
		return err
	}

	content := "工单已转派"
	if remark != "" {
		content += "：" + remark
	}
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "transferred",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Suspend(ticketID, operatorID int64, reason string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status != model.StatusProcessing && ticket.Status != model.StatusAssigned {
		return errors.New("当前状态不能挂起")
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusSuspended); err != nil {
		return err
	}

	content := "工单已挂起"
	if reason != "" {
		content += "：" + reason
	}
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "suspended",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Resume(ticketID, operatorID int64) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status != model.StatusSuspended {
		return errors.New("只有挂起状态的工单可以恢复")
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusProcessing); err != nil {
		return err
	}

	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "resumed",
		OperatorID: operatorID,
		Content:    "工单已恢复处理",
	})
}

func (s *TicketService) AddProgress(ticketID, operatorID int64, content, status string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if status != "" && status == string(model.StatusProcessing) && ticket.Status == model.StatusAssigned {
		_ = s.ticketRepo.UpdateStatus(ticketID, model.StatusProcessing)
	}

	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "progress",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) AddLog(ticketID, operatorID int64, content string) error {
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "log",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Complete(ticketID, operatorID int64, solution, remark string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status != model.StatusProcessing && ticket.Status != model.StatusAssigned {
		return errors.New("当前状态不能完单")
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusReview); err != nil {
		return err
	}

	content := "工单已提交验收，解决方案：" + solution
	if remark != "" {
		content += "，备注：" + remark
	}
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "completed",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Review(ticketID, operatorID int64, approved bool, remark string) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status != model.StatusReview {
		return errors.New("只有待验收状态的工单可以审核")
	}

	if approved {
		if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusCompleted); err != nil {
			return err
		}
		return s.ticketRepo.AddLog(&model.TicketLog{
			TicketID:   ticketID,
			Action:     "reviewed",
			OperatorID: operatorID,
			Content:    "验收通过",
		})
	}

	// Rejected
	if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusProcessing); err != nil {
		return err
	}
	content := "验收驳回"
	if remark != "" {
		content += "：" + remark
	}
	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "rejected",
		OperatorID: operatorID,
		Content:    content,
	})
}

func (s *TicketService) Archive(ticketID, operatorID int64) error {
	ticket, err := s.ticketRepo.FindByID(ticketID)
	if err != nil {
		return errors.New("工单不存在")
	}

	if ticket.Status != model.StatusCompleted {
		return errors.New("只有已完单的工单可以归档")
	}

	if err := s.ticketRepo.UpdateStatus(ticketID, model.StatusArchived); err != nil {
		return err
	}

	return s.ticketRepo.AddLog(&model.TicketLog{
		TicketID:   ticketID,
		Action:     "archived",
		OperatorID: operatorID,
		Content:    "工单已归档",
	})
}

func (s *TicketService) GetLogs(ticketID int64) ([]model.TicketLog, error) {
	return s.ticketRepo.GetLogs(ticketID)
}

func (s *TicketService) Update(ticket *model.Ticket) error {
	return s.ticketRepo.Update(ticket)
}

func (s *TicketService) Delete(ticketID int64) error {
	if _, err := s.ticketRepo.FindByID(ticketID); err != nil {
		return errors.New("工单不存在")
	}
	// Clean up associated files
	if s.completionSvc != nil {
		s.completionSvc.DeleteAllFiles(ticketID)
	}
	return s.ticketRepo.Delete(ticketID)
}
