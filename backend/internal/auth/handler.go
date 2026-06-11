package auth

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/internal/user/userapi"
	"github.com/yourorg/goapp/pkg/apperror"
)

const refreshCookieName = "refresh_token"

// Handler exposes the authentication HTTP endpoints.
type Handler struct {
	svc        *Service
	users      Users
	notifier   Notifier
	cookie     config.CookieConfig
	refreshTTL time.Duration
	logger     *slog.Logger
}

func NewHandler(svc *Service, users Users, notifier Notifier, cookie config.CookieConfig, refreshTTL time.Duration, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, users: users, notifier: notifier, cookie: cookie, refreshTTL: refreshTTL, logger: logger}
}

func (h *Handler) Register(c *gin.Context) {
	req, err := httpx.DecodeValid[RegisterRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	u, err := h.svc.Register(c.Request.Context(), strings.ToLower(strings.TrimSpace(req.Email)), req.Password, req.FirstName, req.LastName)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	// Best-effort async side effect: a failed notification must not fail
	// registration. The notifications module turns this into a welcome email.
	if h.notifier != nil {
		if err := h.notifier.NotifyWelcome(c.Request.Context(), u.ID, u.Email, u.FirstName); err != nil {
			h.logger.Warn("notify welcome", slog.String("err", err.Error()))
		}
	}
	httpx.Data(c, http.StatusCreated, userapi.ToResponse(u))
}

func (h *Handler) Login(c *gin.Context) {
	req, err := httpx.DecodeValid[LoginRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	_, pair, err := h.svc.Login(c.Request.Context(), strings.ToLower(strings.TrimSpace(req.Email)), req.Password)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	httpx.JSON(c, http.StatusOK, TokenResponse{AccessToken: pair.AccessToken, TokenType: "bearer"})
}

func (h *Handler) Refresh(c *gin.Context) {
	token, err := c.Cookie(refreshCookieName)
	if err != nil || token == "" {
		httpx.WriteError(c, h.logger, apperror.Unauthorized("missing_token", "Refresh token is missing"))
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), token)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	h.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	httpx.JSON(c, http.StatusOK, TokenResponse{AccessToken: pair.AccessToken, TokenType: "bearer"})
}

func (h *Handler) Logout(c *gin.Context) {
	if token, err := c.Cookie(refreshCookieName); err == nil {
		_ = h.svc.Logout(c.Request.Context(), token)
	}
	h.clearRefreshCookie(c)
	httpx.JSON(c, http.StatusOK, map[string]string{"detail": "Successfully logged out"})
}

// Me returns the current user with their global permission map.
func (h *Handler) Me(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	u, err := h.users.GetByID(c.Request.Context(), actor.ID)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	resp := userapi.ToResponse(u)
	resp.Permissions = userapi.Permissions(actor, u)
	httpx.Data(c, http.StatusOK, resp)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	actor, _ := authz.ActorFrom(c.Request.Context())
	req, err := httpx.DecodeValid[UpdateProfileRequest](c)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), actor.ID, req.FirstName, req.LastName)
	if err != nil {
		httpx.WriteError(c, h.logger, err)
		return
	}
	httpx.Data(c, http.StatusOK, userapi.ToResponse(u))
}

// ── cookie helpers ───────────────────────────────────────────────────────────

func (h *Handler) setRefreshCookie(c *gin.Context, token string, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		Domain:   h.cookie.Domain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: parseSameSite(h.cookie.SameSite),
	})
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		Domain:   h.cookie.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: parseSameSite(h.cookie.SameSite),
	})
}

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
