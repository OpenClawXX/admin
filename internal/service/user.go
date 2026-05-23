package service

import (
	"crypto/rand"
	"errors"
	"math/big"

	"ops-platform/config"
	"ops-platform/internal/model"
	"ops-platform/internal/pkg/auth"
	"ops-platform/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewUserService(userRepo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{userRepo: userRepo, cfg: cfg}
}

func (s *UserService) Login(username, password string) (*model.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账户已被禁用")
	}

	if !auth.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := auth.GenerateToken(&s.cfg.JWT, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *UserService) Create(req *model.User) error {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return errors.New("密码加密失败")
	}
	req.Password = hashed
	req.Status = 1
	return s.userRepo.Create(req)
}

func (s *UserService) GetByID(id int64) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) List(page, size int, teamID *int64) ([]model.User, int64, error) {
	return s.userRepo.List(page, size, teamID)
}

func (s *UserService) Update(user *model.User) error {
	return s.userRepo.Update(user)
}

func (s *UserService) Delete(id int64) error {
	return s.userRepo.Delete(id)
}

func (s *UserService) ResetPassword(userID int64, newPassword string) (string, error) {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return "", errors.New("用户不存在")
	}

	plainPassword := newPassword
	if plainPassword == "" {
		plainPassword = generateStrongPassword(8)
	}

	hashed, err := auth.HashPassword(plainPassword)
	if err != nil {
		return "", errors.New("密码加密失败")
	}

	if err := s.userRepo.UpdatePassword(userID, hashed); err != nil {
		return "", errors.New("更新密码失败")
	}

	return plainPassword, nil
}

func generateStrongPassword(length int) string {
	const charset = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
