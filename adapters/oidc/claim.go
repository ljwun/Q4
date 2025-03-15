// 參考https://auth0.com/docs/get-started/apis/scopes/openid-connect-scopes
package oidc

import "github.com/coreos/go-oidc/v3/oidc"

type OpenID struct {
	Sub    string `json:"sub"`
	Iss    string `json:"iss"`
	Aud    string `json:"aud"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
	AtHash string `json:"at_hash"`
}

type Email struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type Profile struct {
	Name       string `json:"name"`
	FamilyName string `json:"family_name"`
	GivenName  string `json:"given_name"`
	MiddleName string `json:"middle_name"`
	Nickname   string `json:"nickname"`
	Picture    string `json:"picture"`
	UpdatedAt  string `json:"updated_at"`
}

type IDToken struct {
	OpenID
	Email
	Profile

	internal *oidc.IDToken
}

func (i *IDToken) Claims(v any) error {
	return i.internal.Claims(v)
}
