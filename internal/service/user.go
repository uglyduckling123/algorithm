package service

import (
	"context"
	"errors"
	"time"

	"mango/internal/model"
	"mango/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务接口
type UserService interface {
	Service
	GetByID(ctx context.Context, id uint) (*model.User, error)
	Regist(ctx context.Context, username, email, password string) (*model.User, error)
	Login(ctx context.Context, username, password string) (*model.User, error)
	UpdateProfile(ctx context.Context, id uint, username, email string) (*model.User, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error)
}

// userService 用户服务实现
type UserServiceS struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository) UserService {
	return &UserServiceS{userRepo: userRepo}
}

func (s *UserServiceS) Close() error {
	return s.userRepo.Close()
}

func (s *UserServiceS) GetByID(ctx context.Context, id uint) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserServiceS) Regist(ctx context.Context, username, email, password string) (*model.User, error) {
	// 检查用户名是否已存在
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	_, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserServiceS) Login(ctx context.Context, username, password string) (*model.User, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

func (s *UserServiceS) UpdateProfile(ctx context.Context, id uint, username, email string) (*model.User, error) {
	// 获取用户
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新用户信息
	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserServiceS) DeleteUser(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *UserServiceS) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, offset, pageSize)
}
