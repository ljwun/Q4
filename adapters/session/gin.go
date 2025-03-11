package session

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	SessionKeyForContext = "q4-default-session-context"
	SessionKeyForCookie  = "session"
)

func DefaultGinMiddleware(store IStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 cookie 中取得 session id
		sessionID, err := c.Cookie(SessionKeyForCookie)
		// 如果沒有 cookie 或 cookie 中沒有 session id，則產生一個新的 session id
		if err != nil || sessionID == "" {
			sessionID = uuid.New().String()
		}
		// 建立一個新的 session，並將其加入到 context 中
		c.Set(SessionKeyForContext, NewSession(c.Request.Context(), sessionID, store))
		c.Next()
		// 更新 cookie 中的 session id
		c.SetCookie(SessionKeyForCookie, sessionID, 60*60, "/", "", true, true)
	}
}

func GetDefaultSession(ctx context.Context) (ISession, error) {
	v := ctx.Value(SessionKeyForContext)
	if v == nil {
		return nil, nil
	}
	session := v.(ISession)
	if err := session.Load(); err != nil {
		return nil, err
	}
	return session, nil
}
