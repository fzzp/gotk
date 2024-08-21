package token

import (
	"errors"
)

var ErrInvalidToken = errors.New("令牌无效")

// Maker 定义 jwt 接口两个核心接口
type Maker interface {
	// GenToken 根据用户id生成有时效的token
	GenToken(*Payload) (string, error)

	// ParseToken 解析并验证token是否有效
	ParseToken(token string) (*Payload, error)
}
