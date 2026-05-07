package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

type InvitationHandler struct {
	renderer      *Renderer
	invitationSvc *service.InvitationService
}

func NewInvitationHandler(renderer *Renderer, invitationSvc *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		renderer:      renderer,
		invitationSvc: invitationSvc,
	}
}

type invitationFormData struct {
	Token string
	Name  string
	Error string
}

func (h *InvitationHandler) ShowAcceptForm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	_, user, err := h.invitationSvc.ValidateToken(r.Context(), token)
	if err != nil {
		h.renderInvitationError(w, err)
		return
	}

	h.renderer.HTML(w, http.StatusOK, "invitation/accept.html", &invitationFormData{
		Token: token,
		Name:  user.Name,
	})
}

func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	if password != confirm {
		_, user, _ := h.invitationSvc.ValidateToken(r.Context(), token)
		name := ""
		if user != nil {
			name = user.Name
		}
		h.renderer.HTML(w, http.StatusUnprocessableEntity, "invitation/accept.html", &invitationFormData{
			Token: token,
			Name:  name,
			Error: "パスワードが一致しません",
		})
		return
	}

	if err := h.invitationSvc.AcceptInvitation(r.Context(), token, password); err != nil {
		h.renderInvitationError(w, err)
		return
	}

	http.Redirect(w, r, "/login?invited=1", http.StatusSeeOther)
}

func (h *InvitationHandler) renderInvitationError(w http.ResponseWriter, err error) {
	var msg string
	switch {
	case errors.Is(err, service.ErrTokenExpired):
		msg = "招待リンクの有効期限が切れています。管理者に再送信をご依頼ください。"
	case errors.Is(err, service.ErrTokenUsed):
		msg = "この招待リンクは既に使用されています。"
	default:
		msg = "招待リンクが無効です。"
	}
	h.renderer.HTML(w, http.StatusBadRequest, "invitation/accept.html", &invitationFormData{
		Error: msg,
	})
}
