package service

import (
	"context"

	"mango/internal/model"
	"mango/internal/repository"
)

// UserService 用户服务接口
type TextRiskLogService interface {
	Service
	ListUsers(ctx context.Context, id uint, page, pageSize int) ([]*model.TextRiskLog, error)
}

func (s *TextRiskLogServiceS) Close() error {
	return s.TextRepo.Close()
}

// userService 用户服务实现
type TextRiskLogServiceS struct {
	TextRepo repository.TextRiskLogRepository
}

// NewUserService 创建用户服务
func NewTextRiskLogService(TextRepo repository.TextRiskLogRepository) TextRiskLogService {
	return &TextRiskLogServiceS{TextRepo: TextRepo}
}

func (s *TextRiskLogServiceS) ListUsers(ctx context.Context, id uint, page, pageSize int) ([]*model.TextRiskLog, error) {
	offset := (page - 1) * pageSize
	return s.TextRepo.List(ctx, id, offset, pageSize)
}
