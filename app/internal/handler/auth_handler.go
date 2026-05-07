package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

type AuthHandler struct {
	renderer   *Renderer
	authSvc    service.AuthServicer
	sessionTTL time.Duration
}

func NewAuthHandler(renderer *Renderer, authSvc service.AuthServicer, sessionTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		renderer:   renderer,
		authSvc:    authSvc,
		sessionTTL: sessionTTL,
	}
}

type loginPageData struct {
	Error string
	Email string
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	h.renderer.HTML(w, http.StatusOK, "auth/login.html", &loginPageData{})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderer.HTML(w, http.StatusBadRequest, "auth/login.html", &loginPageData{Error: "リクエストが不正です"})
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	token, _, err := h.authSvc.Login(r.Context(), email, password, r)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, service.ErrUserDisabled):
			msg = "アカウントが無効化されています。管理者にお問い合わせください。"
		default:
			msg = "メールアドレスまたはパスワードが正しくありません。"
		}
		h.renderer.HTML(w, http.StatusUnauthorized, "auth/login.html", &loginPageData{
			Error: msg,
			Email: email,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     service.SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(service.SessionCookieName)
	if err == nil {
		_ = h.authSvc.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     service.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
