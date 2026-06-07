package auth

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/goapp/internal/platform/authz"
	"github.com/yourorg/goapp/internal/platform/config"
	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/internal/user"
	"github.com/yourorg/goapp/pkg/apperror"
)

const refreshCookieName = "refresh_token"

// Handler exposes the authentication HTTP endpoints.
type Handler struct {
	svc        *Service
	userQ      *user.Queries
	cookie     config.CookieConfig
	refreshTTL time.Duration
	logger     *slog.Logger
}

func NewHandler(svc *Service, userQ *user.Queries, cookie config.CookieConfig, refreshTTL time.Duration, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, userQ: userQ, cookie: cookie, refreshTTL: refreshTTL, logger: logger}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.DecodeValid[RegisterRequest](r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	u, err := h.svc.Register(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)), req.Password, req.FirstName, req.LastName)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	httpx.Data(w, http.StatusCreated, user.ToResponse(u))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.DecodeValid[LoginRequest](r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	_, pair, err := h.svc.Login(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)), req.Password)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken, pair.RefreshExpiresAt)
	httpx.JSON(w, http.StatusOK, TokenResponse{AccessToken: pair.AccessToken, TokenType: "bearer"})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		httpx.WriteError(w, h.logger, apperror.Unauthorized("missing_token", "Refresh token is missing"))
		return
	}
	pair, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken, pair.RefreshExpiresAt)
	httpx.JSON(w, http.StatusOK, TokenResponse{AccessToken: pair.AccessToken, TokenType: "bearer"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearRefreshCookie(w)
	httpx.JSON(w, http.StatusOK, map[string]string{"detail": "Successfully logged out"})
}

// Me returns the current user with their global permission map.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	u, err := h.userQ.GetByID(r.Context(), actor.ID)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	resp := user.ToResponse(u)
	resp.Permissions = user.NewPolicy(actor, u).Permissions()
	httpx.Data(w, http.StatusOK, resp)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	actor, _ := authz.ActorFrom(r.Context())
	req, err := httpx.DecodeValid[UpdateProfileRequest](r)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	u, err := h.svc.UpdateProfile(r.Context(), actor.ID, req.FirstName, req.LastName)
	if err != nil {
		httpx.WriteError(w, h.logger, err)
		return
	}
	httpx.Data(w, http.StatusOK, user.ToResponse(u))
}

// ── cookie helpers ───────────────────────────────────────────────────────────

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
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

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
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
