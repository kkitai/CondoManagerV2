package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kkitai/CondoManagerV2/app/internal/csvexport"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/pagination"
	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

type UserHandler struct {
	renderer      *Renderer
	userSvc       service.UserServicer
	invitationSvc service.InvitationServicer
}

func NewUserHandler(renderer *Renderer, userSvc service.UserServicer, invitationSvc service.InvitationServicer) *UserHandler {
	return &UserHandler{
		renderer:      renderer,
		userSvc:       userSvc,
		invitationSvc: invitationSvc,
	}
}

type userListData struct {
	Users      []*domain.User
	Stats      *domain.UserStats
	Pagination pagination.Params
	Sort       queryparam.SortParams
	Query      userListQuery
	CurrentUser *domain.User
}

type userListQuery struct {
	Search     string
	Status     string
	Role       string
	Department string
}

var allowedUserSortCols = []string{"name", "email", "created_at", "last_login_at", "status", "role"}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	listQuery := userListQuery{
		Search:     q.Get("search"),
		Status:     q.Get("status"),
		Role:       q.Get("role"),
		Department: q.Get("department"),
	}

	pag := pagination.Parse(r)
	sort := queryparam.ParseSort(r, allowedUserSortCols, "created_at")

	params := domain.UserListParams{
		Search:     listQuery.Search,
		Status:     listQuery.Status,
		Role:       listQuery.Role,
		Department: listQuery.Department,
		Page:       pag.Page,
		PerPage:    pag.PerPage,
		SortColumn: sort.Column,
		SortOrder:  string(sort.Order),
	}

	users, total, err := h.userSvc.List(r.Context(), params)
	if err != nil {
		Error(w, http.StatusInternalServerError, "ユーザー一覧の取得に失敗しました")
		return
	}
	pag.SetTotal(total)

	stats, err := h.userSvc.GetStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "統計情報の取得に失敗しました")
		return
	}

	data := &userListData{
		Users:      users,
		Stats:      stats,
		Pagination: pag,
		Sort:       sort,
		Query:      listQuery,
		CurrentUser: middleware.CurrentUser(r),
	}

	if r.Header.Get("HX-Request") == "true" {
		h.renderer.Partial(w, http.StatusOK, "users/_table.html", data)
		return
	}

	h.renderer.HTML(w, http.StatusOK, "users/list.html", data)
}

type userFormData struct {
	User        *domain.User
	Errors      validator.Errors
	IsNew       bool
	CurrentUser *domain.User
}

func (h *UserHandler) New(w http.ResponseWriter, r *http.Request) {
	h.renderer.HTML(w, http.StatusOK, "users/form.html", &userFormData{
		User:        &domain.User{},
		IsNew:       true,
		CurrentUser: middleware.CurrentUser(r),
	})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	input := service.CreateUserInput{
		Email:      r.FormValue("email"),
		Name:       r.FormValue("name"),
		Role:       r.FormValue("role"),
		Department: r.FormValue("department"),
		JobTitle:   r.FormValue("job_title"),
		Password:   r.FormValue("password"),
	}

	_, err := h.userSvc.Create(r.Context(), input)
	if err != nil {
		var ve validator.Errors
		if ok := asValidatorErrors(err, &ve); ok {
			h.renderer.HTML(w, http.StatusUnprocessableEntity, "users/form.html", &userFormData{
				User: &domain.User{
					Email:    input.Email,
					Name:     input.Name,
					Role:     domain.UserRole(input.Role),
				},
				Errors:      ve,
				IsNew:       true,
				CurrentUser: middleware.CurrentUser(r),
			})
			return
		}
		if err == service.ErrEmailAlreadyExists {
			h.renderer.HTML(w, http.StatusUnprocessableEntity, "users/form.html", &userFormData{
				User: &domain.User{Email: input.Email, Name: input.Name, Role: domain.UserRole(input.Role)},
				Errors: validator.Errors{"email": "このメールアドレスは既に使用されています"},
				IsNew:  true,
				CurrentUser: middleware.CurrentUser(r),
			})
			return
		}
		Error(w, http.StatusInternalServerError, "ユーザーの作成に失敗しました")
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *UserHandler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "ユーザーが見つかりません")
		return
	}

	h.renderer.HTML(w, http.StatusOK, "users/show.html", map[string]any{
		"User":        user,
		"CurrentUser": middleware.CurrentUser(r),
	})
}

func (h *UserHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "ユーザーが見つかりません")
		return
	}

	h.renderer.HTML(w, http.StatusOK, "users/form.html", &userFormData{
		User:        user,
		IsNew:       false,
		CurrentUser: middleware.CurrentUser(r),
	})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	input := service.UpdateUserInput{
		Email:      r.FormValue("email"),
		Name:       r.FormValue("name"),
		Role:       r.FormValue("role"),
		Department: r.FormValue("department"),
		JobTitle:   r.FormValue("job_title"),
	}

	user, err := h.userSvc.Update(r.Context(), id, input)
	if err != nil {
		var ve validator.Errors
		if ok := asValidatorErrors(err, &ve); ok {
			u, _ := h.userSvc.GetByID(r.Context(), id)
			if u == nil {
				u = &domain.User{ID: id, Email: input.Email, Name: input.Name, Role: domain.UserRole(input.Role)}
			}
			h.renderer.HTML(w, http.StatusUnprocessableEntity, "users/form.html", &userFormData{
				User:        u,
				Errors:      ve,
				IsNew:       false,
				CurrentUser: middleware.CurrentUser(r),
			})
			return
		}
		Error(w, http.StatusInternalServerError, "ユーザーの更新に失敗しました")
		return
	}

	_ = user
	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)
}

func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	status := r.FormValue("status")
	if err := h.userSvc.UpdateStatus(r.Context(), id, status); err != nil {
		Error(w, http.StatusInternalServerError, "ステータスの更新に失敗しました")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		user, err := h.userSvc.GetByID(r.Context(), id)
		if err != nil {
			Error(w, http.StatusInternalServerError, "ユーザーの取得に失敗しました")
			return
		}
		h.renderer.Partial(w, http.StatusOK, "users/_row.html", user)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *UserHandler) SendInvitation(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	if err := h.invitationSvc.SendInvitation(r.Context(), id); err != nil {
		Error(w, http.StatusInternalServerError, "招待メールの送信に失敗しました: "+err.Error())
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/users/%d", id), http.StatusSeeOther)
}

func (h *UserHandler) Export(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := domain.UserListParams{
		Search:     q.Get("search"),
		Status:     q.Get("status"),
		Role:       q.Get("role"),
		Department: q.Get("department"),
		Page:       1,
		PerPage:    10000,
	}

	users, _, err := h.userSvc.List(r.Context(), params)
	if err != nil {
		Error(w, http.StatusInternalServerError, "エクスポートに失敗しました")
		return
	}

	filename := fmt.Sprintf("users_%s.csv", time.Now().Format("20060102"))
	exp := csvexport.New(w, filename)

	_ = exp.WriteHeader([]string{"ID", "名前", "メールアドレス", "ロール", "部署", "役職", "ステータス", "最終ログイン", "作成日"})
	for _, u := range users {
		lastLogin := ""
		if u.LastLoginAt != nil {
			lastLogin = u.LastLoginAt.Format("2006-01-02 15:04:05")
		}
		dept := ""
		if u.Department != nil {
			dept = *u.Department
		}
		jobTitle := ""
		if u.JobTitle != nil {
			jobTitle = *u.JobTitle
		}
		_ = exp.WriteRow([]string{
			strconv.FormatInt(u.ID, 10),
			u.Name,
			u.Email,
			string(u.Role),
			dept,
			jobTitle,
			string(u.Status),
			lastLogin,
			u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	exp.Flush()
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func asValidatorErrors(err error, ve *validator.Errors) bool {
	errs, ok := err.(validator.Errors)
	if ok {
		*ve = errs
	}
	return ok
}
