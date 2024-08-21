package token

import (
	"errors"
	"time"

	"github.com/pascaldekloe/jwt"
)

const minSecretKeySize = 32

// JWTMaker jwt token 生产/解析结构体
type JWTMaker struct {
	secretKey string // 密钥
	issuer    string // 签发人
}

// NewJWTMaker 创建一个维护Token生成、解析的对象，secretKey：密钥，issuer签发token主体
func NewJWTMaker(secretKey string, issuer string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, errors.New("Token密钥长度必须>=32")
	}
	if issuer == "" {
		return nil, errors.New("Token签发人必填项")
	}

	maker := &JWTMaker{
		secretKey: secretKey,
		issuer:    issuer,
	}

	return maker, nil
}

// GenToken 生成 Token
func (maker *JWTMaker) GenToken(payload *Payload) (string, error) {
	// 声明
	var claims jwt.Claims

	// 唯一标识
	claims.ID = payload.ID
	// 主体唯一标识，可以用于存用户唯一标识，面向用户或者说使用者标识
	claims.Subject = payload.UserText
	// 签发时间
	claims.Issued = jwt.NewNumericTime(payload.IssuedAt)
	// 令牌生效时间
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	// 过期时间
	claims.Expires = jwt.NewNumericTime(payload.ExpiredAt)
	// 签发人
	claims.Issuer = maker.issuer
	// Audience是指令牌的受众，通常，Audience指定为服务端的标识符
	claims.Audiences = []string{maker.issuer}

	// 用密钥签名 JWT
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(maker.secretKey))
	if err != nil {
		return "", err
	}

	return string(jwtBytes), nil
}

// ParseToken 解析并验证 Token
func (maker *JWTMaker) ParseToken(token string) (*Payload, error) {
	claims, err := jwt.HMACCheck([]byte(token), []byte(maker.secretKey))
	if err != nil {
		return nil, err
	}

	if !claims.Valid(time.Now()) || claims.Issuer != maker.issuer || !claims.AcceptAudience(maker.issuer) {
		return nil, ErrInvalidToken
	}

	payload := &Payload{
		ID:        claims.ID,
		UserText:  claims.Subject,
		IssuedAt:  claims.Issued.Time().Local(),
		ExpiredAt: claims.Expires.Time().Local(),
	}

	return payload, nil
}
