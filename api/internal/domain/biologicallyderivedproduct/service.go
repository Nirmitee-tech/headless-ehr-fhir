package biologicallyderivedproduct

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	repo BiologicallyDerivedProductRepository
	vt   *fhir.VersionTracker
}

func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) { s.vt = vt }
func (s *Service) VersionTracker() *fhir.VersionTracker      { return s.vt }

func NewService(repo BiologicallyDerivedProductRepository) *Service {
	return &Service{repo: repo}
}

var validProductCategories = map[string]bool{
	"organ": true, "tissue": true, "fluid": true, "cells": true, "biologicalAgent": true,
}

var validStatuses = map[string]bool{
	"available": true, "unavailable": true,
}

func (s *Service) CreateBiologicallyDerivedProduct(ctx context.Context, b *BiologicallyDerivedProduct) error {
	if b.ProductCategory != nil && !validProductCategories[*b.ProductCategory] {
		return fmt.Errorf("invalid product category: %s", *b.ProductCategory)
	}
	if b.Status != nil && !validStatuses[*b.Status] {
		return fmt.Errorf("invalid status: %s", *b.Status)
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return err
	}
	b.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "BiologicallyDerivedProduct", b.FHIRID, b.ToFHIR())
	}
	return nil
}

func (s *Service) GetBiologicallyDerivedProduct(ctx context.Context, id uuid.UUID) (*BiologicallyDerivedProduct, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBiologicallyDerivedProductByFHIRID(ctx context.Context, fhirID string) (*BiologicallyDerivedProduct, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateBiologicallyDerivedProduct(ctx context.Context, b *BiologicallyDerivedProduct) error {
	if b.ProductCategory != nil && !validProductCategories[*b.ProductCategory] {
		return fmt.Errorf("invalid product category: %s", *b.ProductCategory)
	}
	if b.Status != nil && !validStatuses[*b.Status] {
		return fmt.Errorf("invalid status: %s", *b.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "BiologicallyDerivedProduct", b.FHIRID, b.VersionID, b.ToFHIR())
		if err == nil {
			b.VersionID = newVer
		}
	}
	return s.repo.Update(ctx, b)
}

func (s *Service) DeleteBiologicallyDerivedProduct(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		b, err := s.repo.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "BiologicallyDerivedProduct", b.FHIRID, b.VersionID)
		}
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) SearchBiologicallyDerivedProducts(ctx context.Context, params map[string]string, limit, offset int) ([]*BiologicallyDerivedProduct, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}
