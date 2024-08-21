package token

import (
	"time"

	"github.com/google/uuid"
)

// Payload 定义Token负载数据
type Payload struct {
	ID        string    `json:"id"`        // Token 唯一标识
	UserText  string    `json:"userText"`  // 用户数据
	IssuedAt  time.Time `json:"issuedAt"`  // 签发时间
	ExpiredAt time.Time `json:"expiredAt"` // 过期时间
}

// NewPayload 创建一个 Payload 提供给 GenToken 方法使用
func NewPayload(userText string, duration time.Duration) *Payload {
	payload := &Payload{
		ID:        uuid.NewString(), // 可能会panic
		UserText:  userText,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload
}
