package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ExchangeVerifier 用於驗證 OIDC 身份驗證過程中的令牌和狀態
type ExchangeVerifier struct {
	idTokenVerifier *oidc.IDTokenVerifier // ID 令牌驗證器
	reqState        string                // 請求狀態值
	reqNonce        string                // 請求隨機數
}

// VerifyIDToken 驗證 ID 令牌的有效性
func (v *ExchangeVerifier) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	const op = "VerifyIDToken"
	idToken, err := v.idTokenVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("[%s] err=%w", op, err)
	}
	return idToken, nil
}

// VerifyState 驗證狀態值是否匹配
func (v *ExchangeVerifier) VerifyState(state string) bool {
	return state == v.reqState
}

// VerifyNonce 驗證隨機數是否匹配
func (v *ExchangeVerifier) VerifyNonce(nonce string) bool {
	return nonce == v.reqNonce
}
