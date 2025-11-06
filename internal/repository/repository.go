package repository

// Repository 基础仓库接口
type Repository interface {
	Close() error
}
