package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"ats-backend/config"
	"ats-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxAdminID      = "admin_id"
	CtxCompanyID    = "company_id"
	CtxRole         = "role"
	ctxAdminToken   = "admin_token"
	CtxSuperAdminID = "super_admin_id"
)

// Sessions live in HttpOnly cookies so page scripts (and any XSS) can't read the token.
const (
	AdminCookie      = "ats_session"
	SuperAdminCookie = "ats_super_session"
	// ClientHeader must accompany state-changing requests. Browsers only send custom headers
	// cross-origin after a CORS preflight, which only allowlisted origins pass, so this blocks CSRF.
	ClientHeader = "X-ATS-Client"
	sessionTTL   = 24 * time.Hour
)

// cookieSecure is false only for plain-http local development (COOKIE_SECURE=false).
func cookieSecure() bool { return os.Getenv("COOKIE_SECURE") != "false" }

// SetSessionCookie stores a session token. Secure cookies are SameSite=None + Partitioned so the
// embedded dashboard keeps working inside third-party iframes.
func SetSessionCookie(c *gin.Context, name, token string) {
	writeSessionCookie(c, name, token, int(sessionTTL.Seconds()))
}

func ClearSessionCookie(c *gin.Context, name string) {
	writeSessionCookie(c, name, "", -1)
}

func writeSessionCookie(c *gin.Context, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name: name, Value: value, Path: "/api", MaxAge: maxAge,
		HttpOnly: true, Secure: cookieSecure(), SameSite: http.SameSiteLaxMode,
	}
	if cookie.Secure {
		cookie.SameSite, cookie.Partitioned = http.SameSiteNoneMode, true
	}
	http.SetCookie(c.Writer, cookie)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := sessionToken(c, AdminCookie)
		if !ok {
			return
		}
		claims, err := utils.VerifyAdminToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		adminID, companyID := claims.AdminID, claims.CompanyID
		role, active, err := sessionActive(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Could not verify session"})
			return
		}
		if !active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Account no longer active"})
			return
		}
		c.Set(CtxAdminID, adminID)
		c.Set(CtxCompanyID, companyID)
		c.Set(CtxRole, role)
		c.Set(ctxAdminToken, claims)
		c.Next()
	}
}

// SessionStillActive re-runs the session check (cached) for long-lived requests such as event
// streams, so a removed admin or suspended company is cut off mid-stream too.
func SessionStillActive(c *gin.Context) bool {
	claims, ok := c.Get(ctxAdminToken)
	if !ok {
		return false
	}
	t := claims.(utils.AdminToken)
	if time.Now().After(t.IssuedAt.Add(sessionTTL)) {
		return false
	}
	_, active, err := sessionActive(t)
	return err == nil && active
}

// RequireOwner limits a route to company owners; use after AuthMiddleware.
func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(CtxRole) != "owner" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Only account owners can do this"})
			return
		}
		c.Next()
	}
}

// A valid token isn't enough: the admin may have been deleted, the company suspended, or the
// password changed since it was issued. Results are cached briefly so this isn't a query per request.
const sessionCacheTTL = 30 * time.Second

type sessionEntry struct {
	companyID uuid.UUID
	role      string
	active    bool
	expires   time.Time
}

type sessionKey struct {
	adminID  uuid.UUID
	issuedAt int64
}

var sessionCache sync.Map // sessionKey -> sessionEntry

// sessionActive also returns the admin's current role, so role changes apply without re-login.
func sessionActive(t utils.AdminToken) (string, bool, error) {
	key := sessionKey{t.AdminID, t.IssuedAt.Unix()}
	if v, ok := sessionCache.Load(key); ok {
		if e := v.(sessionEntry); time.Now().Before(e.expires) && e.companyID == t.CompanyID {
			return e.role, e.active, nil
		}
	}
	// JWT iat has second precision, so the password change time is truncated to match.
	var rows []struct{ Role string }
	err := config.DB.Raw(`select a.role from admins a join companies c on c.id = a.company_id
		where a.id = ? and a.company_id = ? and c.subscription_status <> 'suspended'
		  and (a.password_changed_at is null or date_trunc('second', a.password_changed_at) <= ?)`,
		t.AdminID, t.CompanyID, t.IssuedAt).Scan(&rows).Error
	if err != nil {
		return "", false, err
	}
	role, active := "", len(rows) == 1
	if active {
		role = rows[0].Role
	}
	sessionCache.Store(key, sessionEntry{t.CompanyID, role, active, time.Now().Add(sessionCacheTTL)})
	return role, active, nil
}

// ForgetAdminSessions drops cached checks for one admin, e.g. after a password reset.
func ForgetAdminSessions(adminID uuid.UUID) {
	sessionCache.Range(func(key, _ any) bool {
		if key.(sessionKey).adminID == adminID {
			sessionCache.Delete(key)
		}
		return true
	})
}

// ForgetCompanySessions drops cached checks so a suspension takes effect immediately.
func ForgetCompanySessions(companyID uuid.UUID) {
	sessionCache.Range(func(key, value any) bool {
		if value.(sessionEntry).companyID == companyID {
			sessionCache.Delete(key)
		}
		return true
	})
}

func SuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := sessionToken(c, SuperAdminCookie)
		if !ok {
			return
		}
		superAdminID, err := utils.VerifySuperAdminJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set(CtxSuperAdminID, superAdminID)
		c.Next()
	}
}

// sessionToken enforces the CSRF header on writes, then returns the session cookie's token.
func sessionToken(c *gin.Context, cookieName string) (string, bool) {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		if c.GetHeader(ClientHeader) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Missing " + ClientHeader + " header"})
			return "", false
		}
	}
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not signed in"})
		return "", false
	}
	return token, true
}
