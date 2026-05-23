package service

import (
	"ops-platform/internal/model"
	"ops-platform/internal/repository"
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) Create(project *model.Project, memberIDs []int64) error {
	if project.Status == "" {
		project.Status = "active"
	}
	code, err := s.projectRepo.NextCode()
	if err != nil {
		return err
	}
	project.Code = code

	if err := s.projectRepo.Create(project); err != nil {
		return err
	}

	if len(memberIDs) > 0 {
		_ = s.projectRepo.SetMembers(project.ID, memberIDs)
	}
	return nil
}

func (s *ProjectService) GetByID(id int64) (*model.Project, error) {
	return s.projectRepo.FindByID(id)
}

func (s *ProjectService) List() ([]model.Project, error) {
	return s.projectRepo.List()
}

func (s *ProjectService) Update(project *model.Project, memberIDs []int64) error {
	if err := s.projectRepo.Update(project); err != nil {
		return err
	}
	if memberIDs != nil {
		_ = s.projectRepo.SetMembers(project.ID, memberIDs)
	}
	return nil
}

func (s *ProjectService) Delete(id int64) error {
	return s.projectRepo.Delete(id)
}

func (s *ProjectService) GetMemberIDs(projectID int64) ([]int64, error) {
	return s.projectRepo.GetMemberIDs(projectID)
}
