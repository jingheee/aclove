package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/jsonutil"
	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/session"
)

const (
	ContextUserKey = "anonymous_user"
	ContextSession = "session"
)

type SessionMiddleware struct {
	manager *session.Manager
}

func NewSessionMiddleware(manager *session.Manager) *SessionMiddleware {
	return &SessionMiddleware{manager: manager}
}

func (m *SessionMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieName, _, cookieDomain, secure, httpOnly, _ := m.manager.CookieConfig()

		sessionID, err := c.Cookie(cookieName)
		if err != nil && err != http.ErrNoCookie {
			logger.Warn("读取cookie失败", "error", err, "ip", c.ClientIP())
		}

		fingerprint := c.GetHeader("X-Device-Fingerprint")
		clientIP := ExtractClientIP(c.Request.RemoteAddr)

		sess, err := m.manager.GetOrCreateSession(c.Request.Context(), sessionID, clientIP, fingerprint)
		if err != nil {
			switch e := err.(type) {
		case *session.BannedError:
			logger.Warn("被封禁用户访问", "ip", clientIP, "until", e.BannedUntil)
			jsonutil.JSON403(c, gin.H{
				"error":        "用户已被封禁",
				"reason":       e.Reason,
				"banned_until": e.BannedUntil,
			})
			c.Abort()
			return
		case *session.CooldownError:
			logger.Warn("冷却期用户访问", "ip", clientIP, "until", e.CooldownUntil)
			jsonutil.JSON(c, http.StatusTooManyRequests, gin.H{
				"error":          "用户处于冷却期",
				"cooldown_until": e.CooldownUntil,
			})
			c.Abort()
			return
		default:
			logger.Error("获取session失败", "error", err, "ip", clientIP)
			jsonutil.JSON500(c, gin.H{"error": "服务暂时不可用"})
			c.Abort()
			return
		}
		}

		if sessionID == "" || sessionID != sess.ID {
			_, maxAge, _, _, _, _ := m.manager.CookieConfig()
			c.SetCookie(
				cookieName,
				sess.ID,
				maxAge,
				"/",
				cookieDomain,
				secure,
				httpOnly,
			)
			logger.Debug("设置新cookie", "user_id", sess.UserID, "ip", clientIP)
		}

		c.Set(ContextSession, sess)
		if sess.UserInfo != nil {
			c.Set(ContextUserKey, sess.UserInfo)
		}

		c.Next()
	}
}

func GetSession(c *gin.Context) (*session.Session, bool) {
	sess, exists := c.Get(ContextSession)
	if !exists {
		return nil, false
	}
	s, ok := sess.(*session.Session)
	return s, ok
}

func MustGetSession(c *gin.Context) *session.Session {
	sess, exists := GetSession(c)
	if !exists {
		panic("session不存在，请确保使用了Session中间件")
	}
	return sess
}

func GetAnonymousUser(c *gin.Context) (*session.UserInfo, bool) {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}
	u, ok := user.(*session.UserInfo)
	return u, ok
}

func MustGetAnonymousUser(c *gin.Context) *session.UserInfo {
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
	return int64(user.ID)
}

func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetAnonymousUser(c)
	return exists
}

func OptionalSession(manager *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieName, _, _, _, _, _ := manager.CookieConfig()

		sessionID, err := c.Cookie(cookieName)
		if err == http.ErrNoCookie {
			c.Next()
			return
		}

		if err != nil {
			c.Next()
			return
		}

		fingerprint := c.GetHeader("X-Device-Fingerprint")
		clientIP := ExtractClientIP(c.Request.RemoteAddr)

		sess, err := manager.GetOrCreateSession(c.Request.Context(), sessionID, clientIP, fingerprint)
		if err != nil {
			c.Next()
			return
		}

		c.Set(ContextSession, sess)
		if sess.UserInfo != nil {
			c.Set(ContextUserKey, sess.UserInfo)
		}

		c.Next()
	}
}

func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetAnonymousUser(c)
		if !exists {
			jsonutil.JSON401(c, gin.H{"error": "请先访问获取会话"})
			c.Abort()
			return
		}

		if user.IsBanned() {
			jsonutil.JSON403(c, gin.H{"error": "用户已被封禁"})
			c.Abort()
			return
		}

		if user.IsCooldown() {
			jsonutil.JSON(c, http.StatusTooManyRequests, gin.H{
				"error":          "用户处于冷却期",
				"cooldown_until": user.CooldownUntil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ExtractClientIP(remoteAddr string) string {
	for i := len(remoteAddr) - 1; i >= 0; i-- {
		if remoteAddr[i] == ':' {
			return remoteAddr[:i]
		}
	}
	return remoteAddr
}
