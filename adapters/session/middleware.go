package session

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DefaultSessionKeyForContext = "q4-default-session-context"
)

var ErrSessionNotFound = fmt.Errorf("session not found")

// MiddlewareOptions 包含所有 session middleware 的設定選項
type MiddlewareOptions struct {
	sessionKeyForCookie  string        // session 在 cookie 中的 key
	sessionKeyForContext string        // session 在 context 中的 key
	cookieMaxAge         time.Duration // cookie 的過期時間
	cookiePath           string        // cookie 的路徑
	cookieDomain         string        // cookie 的域名
	cookieSecure         bool          // 是否只在 HTTPS 連線中傳送 cookie
	cookieHTTPOnly       bool          // 是否禁止 JavaScript 訪問 cookie
	cookieSameSite       string        // cookie 的 SameSite 屬性
}

// MiddlewareOption 定義設定選項的函數類型
type MiddlewareOption func(*MiddlewareOptions)

// WithSessionKeyForCookie 設定 session 在 cookie 中的 key
func WithSessionKeyForCookie(key string) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.sessionKeyForCookie = key
	}
}

// WithSessionKeyForContext 設定 session 在 context 中的 key
func WithSessionKeyForContext(key string) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.sessionKeyForContext = key
	}
}

// WithCookieMaxAge 設定 cookie 的過期時間
func WithCookieMaxAge(maxAge time.Duration) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookieMaxAge = maxAge
	}
}

// WithCookiePath 設定 cookie 的路徑
func WithCookiePath(path string) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookiePath = path
	}
}

// WithCookieDomain 設定 cookie 的域名
func WithCookieDomain(domain string) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookieDomain = domain
	}
}

// WithCookieSecure 設定是否只在 HTTPS 連線中傳送 cookie
func WithCookieSecure(secure bool) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookieSecure = secure
	}
}

// WithCookieHTTPOnly 設定是否禁止 JavaScript 訪問 cookie
func WithCookieHTTPOnly(httpOnly bool) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookieHTTPOnly = httpOnly
	}
}

// WithCookieSameSite 設定 cookie 的 SameSite 屬性
func WithCookieSameSite(sameSite string) MiddlewareOption {
	return func(options *MiddlewareOptions) {
		options.cookieSameSite = sameSite
	}
}

// GinMiddleware 建立一個 gin 的 session middleware
func GinMiddleware(store IStore, opts ...MiddlewareOption) gin.HandlerFunc {
	// 設定預設選項
	options := MiddlewareOptions{
		sessionKeyForCookie:  "session",
		sessionKeyForContext: DefaultSessionKeyForContext,
		cookieMaxAge:         24 * time.Hour,
		cookiePath:           "/",
		cookieDomain:         "",
		cookieSecure:         true,
		cookieHTTPOnly:       true,
	}

	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}

	return func(c *gin.Context) {
		// 從 cookie 中取得 session id
		sessionID, err := c.Cookie(options.sessionKeyForCookie)
		if err != nil || sessionID == "" {
			sessionID = uuid.New().String()
		}

		// 建立新的 session 並加入 context
		session := NewSession(c.Request.Context(), sessionID, store)
		c.Set(options.sessionKeyForContext, session)

		c.Next()

		// 更新 cookie
		c.SetCookie(
			options.sessionKeyForCookie,
			sessionID,
			int(options.cookieMaxAge/time.Second),
			options.cookiePath,
			options.cookieDomain,
			options.cookieSecure,
			options.cookieHTTPOnly,
		)
	}
}

// GetSession 從 context 中取得 session
func GetSession(ctx context.Context, opts ...MiddlewareOption) (ISession, error) {
	const op = "session.GetSession"
	// 設定預設選項
	options := MiddlewareOptions{
		sessionKeyForContext: DefaultSessionKeyForContext,
	}
	// 應用自定義選項
	for _, opt := range opts {
		opt(&options)
	}
	// 從 context 中取得 session
	v := ctx.Value(options.sessionKeyForContext)
	if v == nil {
		return nil, ErrSessionNotFound
	}
	session, ok := v.(ISession)
	if !ok {
		return nil, fmt.Errorf("%s: invalid session type in context", op)
	}
	// 載入 session 資料
	if err := session.Load(); err != nil {
		return nil, fmt.Errorf("%s: failed to load session: %w", op, err)
	}

	return session, nil
}
