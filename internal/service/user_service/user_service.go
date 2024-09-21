package user_service

import (
	"errors"
	"fadacontrol/internal/entity"
	"fadacontrol/pkg/secure"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}
func (s *UserService) Login(username, password string) (*entity.User, error) {
	var user entity.User
	s.db.Model(&entity.User{}).First(&user, "username = ?", username)

	if user.Username == "" {
		return nil, errors.New("username is empty")
	}
	if user.Password == "" {
		return nil, errors.New("password is empty")
	}
	// 验证密码
	if secure.VerifyPassword(password, user.Salt, user.Password) {
		return &user, nil
	}

	return &user, errors.New("password is wrong")
}
