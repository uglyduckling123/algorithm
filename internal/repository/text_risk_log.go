package repository

import (
	"context"

	"mango/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type TextRiskLogRepository interface {
	Repository
	List(ctx context.Context, id uint, offset, limit int) ([]*model.TextRiskLog, error)
}

// userRepository 用户仓库实现
type TextRiskLogRepositoryS struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewTextRiskLogRepository(db *gorm.DB) TextRiskLogRepository {
	return &TextRiskLogRepositoryS{db: db}
}

func (r *TextRiskLogRepositoryS) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *TextRiskLogRepositoryS) List(ctx context.Context, id uint, offset, limit int) ([]*model.TextRiskLog, error) {
	var users []*model.TextRiskLog
	if err := r.db.WithContext(ctx).Where("status = ?", 1).Where("text != ?", "").Where("id >?", id).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
