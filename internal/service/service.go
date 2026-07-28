// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"errors"
	"sync"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/sanitizer"
)

var ErrValidation = errors.New("validation error")

type AddressRepository interface {
	InsertAddressRequest(ctx context.Context, rec *database.AddressRecord) error
	Ping(ctx context.Context) error
}

type LocationRepository interface {
	Ping(ctx context.Context) error
	FindAllProvinces(ctx context.Context) ([]database.ProvinceRow, error)
	FindAllCities(ctx context.Context) ([]database.CityRow, error)
	FindAllDistricts(ctx context.Context, sourceID int64) ([]database.DistrictRow, error)
	FindAllSubDistricts(ctx context.Context, sourceID int64) ([]database.SubDistrictRow, error)
	FindByPostalCode(ctx context.Context, postalCode string, sourceID int64) ([]model.Location, error)
	FindSourceByCode(ctx context.Context, code string) (int64, string, error)
	LoadCityProvinceMapping(ctx context.Context, sourceID int64) (map[int64]int64, error)
	LoadFullHierarchy(ctx context.Context, sourceID int64) (*database.HierarchyMap, error)
}

type provinceEntry struct {
	ID   int64
	Name string
	Kode string
}

type cityEntry struct {
	ID         int64
	Name       string
	Kode       string
	PostalCode string
}

type districtEntry struct {
	ID   int64
	Name string
	Kode string
}

type subDistrictEntry struct {
	ID         int64
	Name       string
	Kode       string
	PostalCode string
}

type Service struct {
	repo             AddressRepository
	locationRepo     LocationRepository
	s                *sanitizer.Sanitizer
	maxAddressLength int
	sourceCode       string

	provinceCache       map[string][]*provinceEntry
	provinceOnce        sync.Once
	provinceErr         error
	provinceKodeToEntry map[string]*provinceEntry
	provinceByID        map[int64]*provinceEntry

	cityCache        map[string][]*cityEntry
	cityOnce         sync.Once
	cityErr          error
	cityByID         map[int64]*cityEntry

	districtCache map[string][]*districtEntry
	districtOnce  sync.Once
	districtErr   error
	districtByID  map[int64]*districtEntry

	subDistrictCache map[string][]*subDistrictEntry
	subDistrictOnce  sync.Once
	subDistrictErr   error
	subDistrictByID  map[int64]*subDistrictEntry

	hierarchyCache *database.HierarchyMap
	hierarchyOnce  sync.Once
	hierarchyErr   error

	phraseDict     map[string]map[string][]model.Entity
	phraseDictOnce sync.Once
	phraseDictErr  error
}

func New(repo AddressRepository, locationRepo LocationRepository, s *sanitizer.Sanitizer, maxAddressLength int, sourceCode string) *Service {
	return &Service{repo: repo, locationRepo: locationRepo, s: s, maxAddressLength: maxAddressLength, sourceCode: sourceCode}
}

func (svc *Service) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}

func (svc *Service) ValidateAddress(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	return svc.ValidateAddressV1(ctx, req, requestID)
}
