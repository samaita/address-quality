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
	FindByKode(ctx context.Context, kode string, sourceID int64) (*model.Location, error)
	FindByPostalCode(ctx context.Context, postalCode string, sourceID int64) (*model.Location, error)
	FindProvincesBySourceID(ctx context.Context, sourceID int64) ([]database.ProvinceRow, error)
	FindSourceByCode(ctx context.Context, code string) (int64, string, error)
}

type Service struct {
	repo             AddressRepository
	locationRepo     LocationRepository
	s                *sanitizer.Sanitizer
	maxAddressLength int
	sourceCode       string

	provinceCache map[int64][]database.ProvinceRow
	provinceOnce  sync.Once
	provinceErr   error
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
