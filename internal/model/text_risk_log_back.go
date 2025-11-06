package model

import "time"

type TextRiskLog struct {
	ID        uint      `json:"id" gorm:"column:id"`
	Uuid      string    `json:"uuid" gorm:"column:uuid"`             // 用户uuid
	Text      string    `json:"text" gorm:"column:text"`             // 检测文本
	Status    uint8     `json:"status" gorm:"column:status"`         // 状态 1:失败 2:通过 3命中
	RequestID string    `json:"request_id" gorm:"column:request_id"` // 请求id
	ReqBody   string    `json:"req_body" gorm:"column:req_body"`     // 请求信息
	RepBody   string    `json:"rep_body" gorm:"column:rep_body"`     // 响应信息
	IsModel   uint8     `json:"is_model" gorm:"column:is_model"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (TextRiskLog) TableName() string {
	return "text_risk_log_back"
}
