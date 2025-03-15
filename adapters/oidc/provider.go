package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrStateMismatch = errors.New("state mismatch")
	ErrNonceMismatch = errors.New("nonce mismatch")
)

type Provider struct {
	*oidc.Provider

	clientInfo ProvideClientInfo
}

type ProvideClientInfo struct {
	ID     string
	Secret string
}

// NewProvider
func NewProvider(issuerURL, clientID, clientSecret string) (*Provider, error) {
	const op = "NewProvider"
	provider, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to create provider, err=%w", op, err)
	}
	return &Provider{
		Provider: provider,
		clientInfo: ProvideClientInfo{
			ID:     clientID,
			Secret: clientSecret,
		},
	}, nil
}

func (p *Provider) AuthURL(state, nonce, redirectUrl string, scopes []string, opts ...oauth2.AuthCodeOption) string {
	config := oauth2.Config{
		ClientID:     p.clientInfo.ID,
		ClientSecret: p.clientInfo.Secret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  redirectUrl,
		Scopes:       scopes,
	}
	return config.AuthCodeURL(state, oidc.Nonce(nonce))
}

func (p *Provider) Exchange(ctx context.Context, verifier *ExchangeVerifier, code, state, redirectUrl string) (*ExchangeToken, error) {
	const op = "Exchange"
	if !verifier.VerifyState(state) {
		return nil, ErrStateMismatch
	}
	config := oauth2.Config{
		ClientID:     p.clientInfo.ID,
		ClientSecret: p.clientInfo.Secret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  redirectUrl,
	}
	oauth2Token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("[%s] Failed to exchange token, err=%w", op, err)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("[%s] No id_token field in oauth2 token", op)
	}
	idToken, err := verifier.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("[%s] Failed to verify ID Token, err=%w", op, err)
	}
	if !verifier.VerifyNonce(idToken.Nonce) {
		return nil, ErrNonceMismatch
	}
	token := &ExchangeToken{
		OAuth2Token: oauth2Token,
		IDToken:     IDToken{internal: idToken},
	}
	if err := idToken.Claims(&token.IDToken); err != nil {
		return nil, fmt.Errorf("[%s] Failed to parse ID Token claims, err=%w", op, err)
	}

	return token, nil
}

func (p *Provider) NewExchangeVerifier(reqState, reqNonce string) *ExchangeVerifier {
	return &ExchangeVerifier{
		idTokenVerifier: p.Verifier(&oidc.Config{ClientID: p.clientInfo.ID}),
		reqState:        reqState,
		reqNonce:        reqNonce,
	}
}

func (p *Provider) SendClientAuthRequest(req *http.Request) (*json.RawMessage, error) {
	const op = "SendClientAuthRequest"

	req.SetBasicAuth(p.clientInfo.ID, p.clientInfo.Secret)

	// 發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to send request, err=%w", op, err)
	}
	defer resp.Body.Close()

	// 檢查回應狀態碼
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[%s] Request failed with status code=%d", op, resp.StatusCode)
	}

	// 解析回應
	respBody := new(json.RawMessage)
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, fmt.Errorf("[%s] Fail to decode response body, err=%w", op, err)
	}
	return respBody, nil
}

type ExchangeToken struct {
	OAuth2Token *oauth2.Token
	IDToken     IDToken
}
