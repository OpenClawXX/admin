package service

import (
	"ops-platform/internal/model"
	"ops-platform/internal/repository"
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *TeamService) Create(team *model.Team) error {
	if team.SupervisorID != nil {
		if err := s.reassignSupervisor(*team.SupervisorID, 0); err != nil {
			return err
		}
	}
	if err := s.teamRepo.Create(team); err != nil {
		return err
	}
	if team.SupervisorID != nil {
		_ = s.userRepo.UpdateTeamID(*team.SupervisorID, team.ID)
	}
	return nil
}

func (s *TeamService) GetByID(id int64) (*model.Team, error) {
	return s.teamRepo.FindByID(id)
}

func (s *TeamService) List() ([]model.Team, error) {
	return s.teamRepo.List()
}

func (s *TeamService) Update(team *model.Team) error {
	if team.SupervisorID != nil {
		if err := s.reassignSupervisor(*team.SupervisorID, team.ID); err != nil {
			return err
		}
	}
	if err := s.teamRepo.Update(team); err != nil {
		return err
	}
	if team.SupervisorID != nil {
		_ = s.userRepo.UpdateTeamID(*team.SupervisorID, team.ID)
	}
	return nil
}

func (s *TeamService) Delete(id int64) error {
	return s.teamRepo.Delete(id)
}

// reassignSupervisor clears the old team's supervisor if this person is already supervising another team.
// skipTeamID is the current team being edited (0 for create).
func (s *TeamService) reassignSupervisor(supervisorID, skipTeamID int64) error {
	oldTeam, err := s.teamRepo.FindBySupervisorID(supervisorID)
	if err != nil {
		// No existing team with this supervisor — safe to proceed
		return nil
	}
	if oldTeam.ID == skipTeamID {
		// Same team, no conflict
		return nil
	}
	return s.teamRepo.ClearSupervisor(oldTeam.ID)
}
