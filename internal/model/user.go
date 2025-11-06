package model

import "time"

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"nick_name;not null"`
	Email     string    `json:"email" gorm:"email;not null"`
	Password  string    `json:"-" gorm:"not null"` // 密码不返回给客户端
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
