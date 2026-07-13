package service

import (
	"context"
	"errors"

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
	FindByKode(ctx context.Context, kode string) (*model.Location, error)
	FindByPostalCode(ctx context.Context, postalCode string) (*model.Location, error)
}

type Service struct {
	repo             AddressRepository
	locationRepo     LocationRepository
	s                *sanitizer.Sanitizer
	maxAddressLength int
}

func New(repo AddressRepository, locationRepo LocationRepository, s *sanitizer.Sanitizer, maxAddressLength int) *Service {
	return &Service{repo: repo, locationRepo: locationRepo, s: s, maxAddressLength: maxAddressLength}
}

func (svc *Service) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}

func (svc *Service) ValidateAddress(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	return svc.ValidateAddressV1(ctx, req, requestID)
}
