package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/service"
)

const (
	ContextUserKey = "anonymous_user"
	CookieName     = "aclove_session"
)

type SessionMiddleware struct {
	userService *service.AnonymousUserService
}

func NewSessionMiddleware(userService *service.AnonymousUserService) *SessionMiddleware {
	return &SessionMiddleware{userService: userService}
}

func (m *SessionMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(CookieName)
		if err != nil && err != http.ErrNoCookie {
			logger.Warn("读取cookie失败", "error", err, "ip", c.ClientIP())
		}

		fingerprint := c.GetHeader("X-Device-Fingerprint")
		clientIP := service.ExtractClientIP(c.Request.RemoteAddr)

		user, err := m.userService.GetOrCreateUser(c.Request.Context(), cookie, clientIP, fingerprint)
		if err != nil {
			switch e := err.(type) {
			case *service.UserBannedError:
				logger.Warn("被封禁用户访问", "ip", clientIP, "until", e.BannedUntil)
				c.JSON(http.StatusForbidden, gin.H{
					"error":        "用户已被封禁",
					"reason":       e.Reason,
					"banned_until": e.BannedUntil,
				})
				c.Abort()
				return
			case *service.UserCooldownError:
				logger.Warn("冷却期用户访问", "ip", clientIP, "until", e.CooldownUntil)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":          "用户处于冷却期",
					"cooldown_until": e.CooldownUntil,
				})
				c.Abort()
				return
			default:
				logger.Error("获取用户信息失败", "error", err, "ip", clientIP)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务暂时不可用"})
				c.Abort()
				return
			}
		}

		if cookie == "" || cookie != user.Cookie {
			cookieName, cookieExpiration := service.CookieConfig()
			c.SetCookie(
				cookieName,
				user.Cookie,
				int(cookieExpiration.Seconds()),
				"/",
				"",
				false,
				true,
			)
			logger.Debug("设置新cookie", "user_id", user.ID, "ip", clientIP)
		}

		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func GetAnonymousUser(c *gin.Context) (*service.AnonymousUserInfo, bool) {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}
	info, ok := user.(*service.AnonymousUserInfo)
	return info, ok
}

func MustGetAnonymousUser(c *gin.Context) *service.AnonymousUserInfo {
	user, exists := GetAnonymousUser(c)
	if !exists {
		panic("匿名用户信息不存在，请确保使用了Session中间件")
	}
	return user
}

func GetUserID(c *gin.Context) int64 {
	user, exists := GetAnonymousUser(c)
	if !exists {
		return 0
	}
	return user.ID
}

func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetAnonymousUser(c)
	return exists
}

func OptionalSession(userService *service.AnonymousUserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(CookieName)
		if err == http.ErrNoCookie {
			c.Next()
			return
		}

		if err != nil {
			c.Next()
			return
		}

		fingerprint := c.GetHeader("X-Device-Fingerprint")
		clientIP := service.ExtractClientIP(c.Request.RemoteAddr)

		user, err := userService.GetOrCreateUser(c.Request.Context(), cookie, clientIP, fingerprint)
		if err != nil {
			c.Next()
			return
		}

		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetAnonymousUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先访问获取会话"})
			c.Abort()
			return
		}

		if user.Status == service.StatusBanned {
			c.JSON(http.StatusForbidden, gin.H{"error": "用户已被封禁"})
			c.Abort()
			return
		}

		if user.Status == service.StatusCooldown {
			if user.CooldownUntil != nil && user.CooldownUntil.After(time.Now()) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":          "用户处于冷却期",
					"cooldown_until": user.CooldownUntil,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
