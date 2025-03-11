package api

import (
	"q4/adapters/redis"
	"q4/adapters/session"

	"github.com/gin-gonic/gin"
)

const (
	SESSION_KEY_REQUEST_STATE    = "request_state"
	SESSION_KEY_REQUEST_NONCE    = "request_nonce"
	SESSION_KEY_REDIRECT_URL     = "redirect_url"
	SESSION_KEY_URL_BEFORE_LOGIN = "url_before_login"
)

func (impl *ServerImpl) SessionMiddleware() gin.HandlerFunc {
	store := redis.NewStore(
		impl.redisClient,
		redis.WithStorePrefix(impl.config.Redis.KeyPrefix+"session:"),
	)
	return session.GinMiddleware(
		store,
		session.WithSessionKeyForCookie(impl.config.Session.KeyForCookie),
		session.WithCookieMaxAge(impl.config.Session.CookieMaxAge),
	)
}
