package service

// Service 基础服务接口
type Service interface {
	Close() error
}
