package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

var ErrPropertyNotFound = errors.New("property not found")

type PropertyService struct {
	propertyRepo *repository.PropertyRepository
}

func NewPropertyService(propertyRepo *repository.PropertyRepository) *PropertyService {
	return &PropertyService{propertyRepo: propertyRepo}
}

type CreatePropertyInput struct {
	Name              string
	Address           string
	Area              string
	UnitCount         string
	Status            string
	ManagementCompany string
	AssigneeID        int64
	CreatedBy         int64
}

type UpdatePropertyInput struct {
	Name              string
	Address           string
	Area              string
	UnitCount         string
	Status            string
	ManagementCompany string
	AssigneeID        int64
	UpdatedBy         int64
}

func (s *PropertyService) Create(ctx context.Context, in CreatePropertyInput) (*domain.Property, error) {
	v := validator.New()
	v.Required("name", in.Name)
	v.MaxLength("name", in.Name, 200)
	v.Required("address", in.Address)
	v.OneOf("status", in.Status, []string{"active", "inactive"})
	if !v.Valid() {
		return nil, v.Errors()
	}

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
	if in.CreatedBy > 0 {
		p.CreatedBy = &in.CreatedBy
		p.UpdatedBy = &in.CreatedBy
	}

	area, err := parseOptionalFloat(in.Area)
	if err != nil {
		return nil, fmt.Errorf("invalid area: %w", err)
	}
	p.Area = area

	units, err := parseOptionalInt(in.UnitCount)
	if err != nil {
		return nil, fmt.Errorf("invalid unit_count: %w", err)
	}
	p.UnitCount = units

	return s.propertyRepo.Create(ctx, p)
}

func (s *PropertyService) Update(ctx context.Context, id int64, in UpdatePropertyInput) (*domain.Property, error) {
	v := validator.New()
	v.Required("name", in.Name)
	v.MaxLength("name", in.Name, 200)
	v.Required("address", in.Address)
	v.OneOf("status", in.Status, []string{"active", "inactive"})
	if !v.Valid() {
		return nil, v.Errors()
	}

	p, err := s.propertyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrPropertyNotFound
	}

	p.Name = in.Name
	p.Address = in.Address
	p.Status = domain.PropertyStatus(in.Status)

	if in.ManagementCompany != "" {
		p.ManagementCompany = &in.ManagementCompany
	} else {
		p.ManagementCompany = nil
	}
	if in.AssigneeID > 0 {
		p.AssigneeID = &in.AssigneeID
	} else {
		p.AssigneeID = nil
	}
	if in.UpdatedBy > 0 {
		p.UpdatedBy = &in.UpdatedBy
	}

	area, err := parseOptionalFloat(in.Area)
	if err != nil {
		return nil, fmt.Errorf("invalid area: %w", err)
	}
	p.Area = area

	units, err := parseOptionalInt(in.UnitCount)
	if err != nil {
		return nil, fmt.Errorf("invalid unit_count: %w", err)
	}
	p.UnitCount = units

	return s.propertyRepo.Update(ctx, p)
}

func (s *PropertyService) Delete(ctx context.Context, id int64) error {
	if _, err := s.propertyRepo.FindByID(ctx, id); err != nil {
		return ErrPropertyNotFound
	}
	return s.propertyRepo.Delete(ctx, id)
}

func (s *PropertyService) GetByID(ctx context.Context, id int64) (*domain.Property, error) {
	p, err := s.propertyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrPropertyNotFound
	}
	return p, nil
}

func (s *PropertyService) List(ctx context.Context, params domain.PropertyListParams) ([]*domain.Property, int64, error) {
	return s.propertyRepo.List(ctx, params)
}

func (s *PropertyService) GetStats(ctx context.Context, propertyID int64) (*domain.PropertyStats, error) {
	return s.propertyRepo.GetStats(ctx, propertyID)
}

func (s *PropertyService) GetListStats(ctx context.Context) (*domain.PropertyListStats, error) {
	return s.propertyRepo.GetListStats(ctx)
}

func parseOptionalFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func parseOptionalInt(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return nil, err
	}
	return &i, nil
}
