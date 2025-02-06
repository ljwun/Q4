package oidc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ExtendedProvider 擴展了標準 OIDC Provider，增加了令牌撤銷和內省功能
type ExtendedProvider struct {
	*Provider
	Extra ExtraData
}

// ExtraData 包含額外的 OIDC 端點信息
type ExtraData struct {
	RevocationEndpoint    string `json:"revocation_endpoint"`    // 令牌撤銷端點
	IntrospectionEndpoint string `json:"introspection_endpoint"` // 令牌內省端點
}

func NewExtendedProvider(issuerURL, clientID, clientSecret string) (*ExtendedProvider, error) {
	const op = "NewExtendedProviderProvider"
	provider, err := NewProvider(issuerURL, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to create extended provider, err=%w", op, err)
	}
	var extra ExtraData
	if err := provider.Claims(&extra); err != nil {
		return nil, fmt.Errorf("[%s] Fail to claim extra data, err=%w", op, err)
	}
	return &ExtendedProvider{
		Provider: provider,
		Extra:    extra,
	}, nil
}

func (p *ExtendedProvider) Revoke(token string) error {
	const op = "Revoke"
	// 建立請求資料
	form := url.Values{}
	form.Set("token", token)
	body := bytes.NewBufferString(form.Encode())

	// 建立 HTTP 請求
	req, err := http.NewRequest("POST", p.Extra.RevocationEndpoint, body)
	if err != nil {
		return fmt.Errorf("[%s] Fail to create revocation request, err=%w", op, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送請求
	_, err = p.SendClientAuthRequest(req)
	if err != nil {
		return fmt.Errorf("[%s] err=%w", op, err)
	}

	return nil
}

// UserInfo 包含用戶信息的結構體
type UserInfo struct {
	Active        bool     `json:"active"`         // 令牌是否有效
	Name          string   `json:"name"`           // 用戶名稱
	Nickname      string   `json:"nickname"`       // 用戶暱稱
	Email         string   `json:"email"`          // 電子郵件
	EmailVerified bool     `json:"email_verified"` // 郵件是否已驗證
	Groups        []string `json:"groups"`         // 用戶所屬群組
	Scope         string   `json:"scope"`          // 權限範圍
}

func (p *ExtendedProvider) Introspect(token string) (*UserInfo, error) {
	const op = "Introspect"
	// 建立請求資料
	form := url.Values{}
	form.Set("token", token)
	body := bytes.NewBufferString(form.Encode())

	// 建立 HTTP 請求
	req, err := http.NewRequest("POST", p.Extra.IntrospectionEndpoint, body)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to create introspection request, err=%w", op, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送請求
	resp, err := p.SendClientAuthRequest(req)
	if err != nil {
		return nil, fmt.Errorf("[%s] err=%w", op, err)
	}

	// 解析回應
	result := new(UserInfo)
	if err := json.Unmarshal(*resp, result); err != nil {
		return nil, fmt.Errorf("[%s] Fail to retrieve active state, err=%w", op, err)
	}
	return result, nil
}
