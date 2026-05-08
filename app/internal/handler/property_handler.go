package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/csvexport"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/pagination"
	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

type PropertyHandler struct {
	renderer    *Renderer
	propertySvc service.PropertyServicer
}

func NewPropertyHandler(renderer *Renderer, propertySvc service.PropertyServicer) *PropertyHandler {
	return &PropertyHandler{renderer: renderer, propertySvc: propertySvc}
}

type propertyListData struct {
	Properties  []*domain.Property
	Pagination  pagination.Params
	Sort        queryparam.SortParams
	Query       propertyListQuery
	ListStats   *domain.PropertyListStats
	CurrentUser *domain.User
}

type propertyListQuery struct {
	Search string
	Status string
}

var allowedPropertySortCols = []string{"name", "address", "status", "unit_count", "created_at", "updated_at"}

func (h *PropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	listQuery := propertyListQuery{
		Search: q.Get("search"),
		Status: q.Get("status"),
	}

	pag := pagination.Parse(r)
	sort := queryparam.ParseSort(r, allowedPropertySortCols, "created_at")

	params := domain.PropertyListParams{
		Search:     listQuery.Search,
		Status:     listQuery.Status,
		Page:       pag.Page,
		PerPage:    pag.PerPage,
		SortColumn: sort.Column,
		SortOrder:  string(sort.Order),
	}

	props, total, err := h.propertySvc.List(r.Context(), params)
	if err != nil {
		Error(w, http.StatusInternalServerError, "物件一覧の取得に失敗しました")
		return
	}
	pag.SetTotal(total)

	listStats, err := h.propertySvc.GetListStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "統計情報の取得に失敗しました")
		return
	}

	data := &propertyListData{
		Properties:  props,
		Pagination:  pag,
		Sort:        sort,
		Query:       listQuery,
		ListStats:   listStats,
		CurrentUser: middleware.CurrentUser(r),
	}

	if r.Header.Get("HX-Request") == "true" {
		h.renderer.Partial(w, http.StatusOK, "properties/_table.html", data)
		return
	}

	h.renderer.HTML(w, http.StatusOK, "properties/list.html", data)
}

type propertyFormData struct {
	Property    *domain.Property
	Errors      validator.Errors
	IsNew       bool
	CurrentUser *domain.User
}

func (h *PropertyHandler) New(w http.ResponseWriter, r *http.Request) {
	h.renderer.HTML(w, http.StatusOK, "properties/form.html", &propertyFormData{
		Property:    &domain.Property{},
		IsNew:       true,
		CurrentUser: middleware.CurrentUser(r),
	})
}

func (h *PropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	cu := middleware.CurrentUser(r)
	var createdBy int64
	if cu != nil {
		createdBy = cu.ID
	}

	input := service.CreatePropertyInput{
		Name:              r.FormValue("name"),
		Address:           r.FormValue("address"),
		Area:              r.FormValue("area"),
		UnitCount:         r.FormValue("unit_count"),
		Status:            r.FormValue("status"),
		ManagementCompany: r.FormValue("management_company"),
		CreatedBy:         createdBy,
	}
	if assigneeStr := r.FormValue("assignee_id"); assigneeStr != "" {
		if id, err := strconv.ParseInt(assigneeStr, 10, 64); err == nil {
			input.AssigneeID = id
		}
	}

	_, err := h.propertySvc.Create(r.Context(), input)
	if err != nil {
		var ve validator.Errors
		if ok := asValidatorErrors(err, &ve); ok {
			h.renderer.HTML(w, http.StatusUnprocessableEntity, "properties/form.html", &propertyFormData{
				Property:    buildPropertyFromInput(input),
				Errors:      ve,
				IsNew:       true,
				CurrentUser: cu,
			})
			return
		}
		Error(w, http.StatusInternalServerError, "物件の作成に失敗しました")
		return
	}

	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func (h *PropertyHandler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	prop, err := h.propertySvc.GetByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "物件が見つかりません")
		return
	}

	stats, err := h.propertySvc.GetStats(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "統計情報の取得に失敗しました")
		return
	}

	if r.Header.Get("HX-Request") == "true" && r.URL.Query().Get("partial") == "stats" {
		h.renderer.Partial(w, http.StatusOK, "properties/_stats.html", map[string]any{
			"Stats": stats,
		})
		return
	}

	h.renderer.HTML(w, http.StatusOK, "properties/detail.html", map[string]any{
		"Property":    prop,
		"Stats":       stats,
		"CurrentUser": middleware.CurrentUser(r),
	})
}

func (h *PropertyHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	prop, err := h.propertySvc.GetByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "物件が見つかりません")
		return
	}

	h.renderer.HTML(w, http.StatusOK, "properties/form.html", &propertyFormData{
		Property:    prop,
		IsNew:       false,
		CurrentUser: middleware.CurrentUser(r),
	})
}

func (h *PropertyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	if err := r.ParseForm(); err != nil {
		Error(w, http.StatusBadRequest, "リクエストが不正です")
		return
	}

	cu := middleware.CurrentUser(r)
	var updatedBy int64
	if cu != nil {
		updatedBy = cu.ID
	}

	input := service.UpdatePropertyInput{
		Name:              r.FormValue("name"),
		Address:           r.FormValue("address"),
		Area:              r.FormValue("area"),
		UnitCount:         r.FormValue("unit_count"),
		Status:            r.FormValue("status"),
		ManagementCompany: r.FormValue("management_company"),
		UpdatedBy:         updatedBy,
	}
	if assigneeStr := r.FormValue("assignee_id"); assigneeStr != "" {
		if aid, err := strconv.ParseInt(assigneeStr, 10, 64); err == nil {
			input.AssigneeID = aid
		}
	}

	_, err = h.propertySvc.Update(r.Context(), id, input)
	if err != nil {
		var ve validator.Errors
		if ok := asValidatorErrors(err, &ve); ok {
			prop, _ := h.propertySvc.GetByID(r.Context(), id)
			if prop == nil {
				prop = &domain.Property{ID: id}
			}
			h.renderer.HTML(w, http.StatusUnprocessableEntity, "properties/form.html", &propertyFormData{
				Property:    prop,
				Errors:      ve,
				IsNew:       false,
				CurrentUser: cu,
			})
			return
		}
		if err == service.ErrPropertyNotFound {
			Error(w, http.StatusNotFound, "物件が見つかりません")
			return
		}
		Error(w, http.StatusInternalServerError, "物件の更新に失敗しました")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/properties/%d", id), http.StatusSeeOther)
}

func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		Error(w, http.StatusBadRequest, "不正なIDです")
		return
	}

	if err := h.propertySvc.Delete(r.Context(), id); err != nil {
		if err == service.ErrPropertyNotFound {
			Error(w, http.StatusNotFound, "物件が見つかりません")
			return
		}
		Error(w, http.StatusInternalServerError, "物件の削除に失敗しました")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func (h *PropertyHandler) Export(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := domain.PropertyListParams{
		Search:  q.Get("search"),
		Status:  q.Get("status"),
		Page:    1,
		PerPage: 10000,
	}

	props, _, err := h.propertySvc.List(r.Context(), params)
	if err != nil {
		Error(w, http.StatusInternalServerError, "エクスポートに失敗しました")
		return
	}

	filename := fmt.Sprintf("properties_%s.csv", time.Now().Format("20060102"))
	exp := csvexport.New(w, filename)

	_ = exp.WriteHeader([]string{"ID", "物件名", "住所", "面積(㎡)", "戸数", "ステータス", "管理会社", "作成日"})
	for _, p := range props {
		area := ""
		if p.Area != nil {
			area = fmt.Sprintf("%.2f", *p.Area)
		}
		units := ""
		if p.UnitCount != nil {
			units = strconv.Itoa(*p.UnitCount)
		}
		mgmt := ""
		if p.ManagementCompany != nil {
			mgmt = *p.ManagementCompany
		}
		_ = exp.WriteRow([]string{
			strconv.FormatInt(p.ID, 10),
			p.Name,
			p.Address,
			area,
			units,
			string(p.Status),
			mgmt,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	exp.Flush()
}

func buildPropertyFromInput(in service.CreatePropertyInput) *domain.Property {
	p := &domain.Property{
		Name:    in.Name,
		Address: in.Address,
		Status:  domain.PropertyStatus(in.Status),
	}
	if in.ManagementCompany != "" {
		p.ManagementCompany = &in.ManagementCompany
	}
	if in.AssigneeID > 0 {
		p.AssigneeID = &in.AssigneeID
	}
	return p
}
