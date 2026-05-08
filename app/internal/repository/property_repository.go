package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
)

type PropertyRepository struct {
	db *pgxpool.Pool
}

func NewPropertyRepository(db *pgxpool.Pool) *PropertyRepository {
	return &PropertyRepository{db: db}
}

const propertySelectCols = `
	p.id, p.name, p.address, p.area, p.unit_count, p.status,
	p.management_company, p.assignee_id, p.created_by, p.updated_by,
	p.created_at, p.updated_at`

func scanProperty(row interface {
	Scan(dest ...any) error
}) (*domain.Property, error) {
	p := &domain.Property{}
	err := row.Scan(
		&p.ID, &p.Name, &p.Address, &p.Area, &p.UnitCount, &p.Status,
		&p.ManagementCompany, &p.AssigneeID, &p.CreatedBy, &p.UpdatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PropertyRepository) FindByID(ctx context.Context, id int64) (*domain.Property, error) {
	q := fmt.Sprintf(`SELECT %s FROM properties p WHERE p.id = $1`, propertySelectCols)
	p, err := scanProperty(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("find property by id: %w", err)
	}
	return p, nil
}

func (r *PropertyRepository) List(ctx context.Context, params domain.PropertyListParams) ([]*domain.Property, int64, error) {
	fb := queryparam.NewFilterBuilder()

	if params.Search != "" {
		idx := fb.NextIndex()
		fb.Add(
			fmt.Sprintf("(p.name ILIKE $%d OR p.address ILIKE $%d OR p.management_company ILIKE $%d)", idx, idx, idx),
			"%"+params.Search+"%",
		)
	}
	if params.Status != "" {
		fb.AddEqual("p.status", params.Status)
	}
	if params.AssigneeID > 0 {
		fb.AddEqual("p.assignee_id", params.AssigneeID)
	}
	if params.UnitCountMin > 0 {
		fb.Add(fmt.Sprintf("p.unit_count >= $%d", fb.NextIndex()), params.UnitCountMin)
	}
	if params.UnitCountMax > 0 {
		fb.Add(fmt.Sprintf("p.unit_count <= $%d", fb.NextIndex()), params.UnitCountMax)
	}
	fb.AddDateRange("p.updated_at", params.UpdatedFrom, params.UpdatedTo)

	where := fb.WhereClause()
	args := fb.Args()

	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM properties p %s`, where)
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count properties: %w", err)
	}

	sortCol := params.SortColumn
	if sortCol == "" {
		sortCol = "p.created_at"
	}
	sortOrder := strings.ToUpper(params.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	allowedCols := map[string]bool{
		"name": true, "address": true, "status": true,
		"unit_count": true, "created_at": true, "updated_at": true,
	}
	if !allowedCols[sortCol] {
		sortCol = "p.created_at"
	} else {
		sortCol = "p." + sortCol
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	limitIdx := fb.NextIndex()
	offsetIdx := limitIdx + 1
	listQ := fmt.Sprintf(`
		SELECT %s
		FROM properties p
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		propertySelectCols, where, sortCol, sortOrder, limitIdx, offsetIdx,
	)
	listArgs := append(args, perPage, offset)

	rows, err := r.db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	var props []*domain.Property
	for rows.Next() {
		p, err := scanProperty(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan property row: %w", err)
		}
		props = append(props, p)
	}
	return props, total, rows.Err()
}

func (r *PropertyRepository) Create(ctx context.Context, p *domain.Property) (*domain.Property, error) {
	const q = `
		INSERT INTO properties (name, address, area, unit_count, status, management_company, assignee_id, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, q,
		p.Name, p.Address, p.Area, p.UnitCount, p.Status,
		p.ManagementCompany, p.AssigneeID, p.CreatedBy, p.UpdatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create property: %w", err)
	}
	return p, nil
}

func (r *PropertyRepository) Update(ctx context.Context, p *domain.Property) (*domain.Property, error) {
	const q = `
		UPDATE properties
		SET name = $1, address = $2, area = $3, unit_count = $4, status = $5,
		    management_company = $6, assignee_id = $7, updated_by = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, q,
		p.Name, p.Address, p.Area, p.UnitCount, p.Status,
		p.ManagementCompany, p.AssigneeID, p.UpdatedBy, p.ID,
	).Scan(&p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update property: %w", err)
	}
	return p, nil
}

func (r *PropertyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM properties WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete property: %w", err)
	}
	return nil
}

func (r *PropertyRepository) GetStats(ctx context.Context, propertyID int64) (*domain.PropertyStats, error) {
	const q = `
		SELECT
			COUNT(*)                                                        AS total,
			COUNT(*) FILTER (WHERE status = 'pending')                     AS open,
			COUNT(*) FILTER (WHERE status = 'in_progress')                 AS in_progress,
			COUNT(*) FILTER (WHERE status = 'completed')                   AS completed,
			AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0)
				FILTER (WHERE completed_at IS NOT NULL)                     AS avg_response_days,
			CASE WHEN COUNT(*) > 0
				THEN COUNT(*) FILTER (WHERE is_recurrence = TRUE)::FLOAT / COUNT(*)
				ELSE NULL
			END                                                             AS recurrence_rate
		FROM claims
		WHERE property_id = $1`

	s := &domain.PropertyStats{}
	err := r.db.QueryRow(ctx, q, propertyID).Scan(
		&s.TotalClaims, &s.OpenClaims, &s.InProgressClaims, &s.CompletedClaims,
		&s.AvgResponseDays, &s.RecurrenceRate,
	)
	if err != nil {
		return nil, fmt.Errorf("get property stats: %w", err)
	}

	catRows, err := r.db.Query(ctx,
		`SELECT COALESCE(category, '未分類'), COUNT(*) FROM claims WHERE property_id = $1 GROUP BY category ORDER BY COUNT(*) DESC`,
		propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get category breakdown: %w", err)
	}
	defer catRows.Close()
	for catRows.Next() {
		var cc domain.CategoryCount
		if err := catRows.Scan(&cc.Category, &cc.Count); err != nil {
			return nil, fmt.Errorf("scan category count: %w", err)
		}
		s.CategoryBreakdown = append(s.CategoryBreakdown, cc)
	}
	if err := catRows.Err(); err != nil {
		return nil, err
	}

	sevRows, err := r.db.Query(ctx,
		`SELECT severity, COUNT(*) FROM claims WHERE property_id = $1 GROUP BY severity ORDER BY COUNT(*) DESC`,
		propertyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get severity breakdown: %w", err)
	}
	defer sevRows.Close()
	for sevRows.Next() {
		var sc domain.SeverityCount
		if err := sevRows.Scan(&sc.Severity, &sc.Count); err != nil {
			return nil, fmt.Errorf("scan severity count: %w", err)
		}
		s.SeverityBreakdown = append(s.SeverityBreakdown, sc)
	}
	return s, sevRows.Err()
}
