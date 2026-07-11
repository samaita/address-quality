package service

import (
	"context"
	"errors"
	"time"

	"encoding/json"

	"github.com/google/uuid"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/sanitizer"
)

var ErrValidation = errors.New("validation error")

type AddressRepository interface {
	InsertAddressRequest(ctx context.Context, rec *database.AddressRecord) error
	Ping(ctx context.Context) error
}

type Service struct {
	repo             AddressRepository
	s                *sanitizer.Sanitizer
	maxAddressLength int
}

func New(repo AddressRepository, s *sanitizer.Sanitizer, maxAddressLength int) *Service {
	return &Service{repo: repo, s: s, maxAddressLength: maxAddressLength}
}

func (svc *Service) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}

func (svc *Service) ValidateAddress(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	if err := req.Validate(svc.maxAddressLength); err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	now := time.Now().UTC()
	addressID := uuid.Must(uuid.NewV7()).String()
	sanitized := svc.s.Sanitize(req.Address)

	quality := model.Quality{
		AddressID:       addressID,
		Confidence:      0.0,
		Location:        model.Location{},
		NormalizedInput: sanitized,
		Output:          sanitized,
		LocationVersion: "",
		RawInput:        req.Address,
	}

	outputJSON, _ := json.Marshal(quality)

	record := &database.AddressRecord{
		ID:              requestID,
		AddressID:       addressID,
		RawInput:        req.Address,
		NormalizedAddr:  sanitized,
		Confidence:      0.0,
		PostalCode:      "",
		SubDistrict:     "",
		District:        "",
		City:            "",
		Province:        "",
		LocationVersion: "",
		OutputJSON:      string(outputJSON),
		CreatedAt:       now,
	}

	if err := svc.repo.InsertAddressRequest(ctx, record); err != nil {
		return nil, err
	}

	resp := &model.AddressResponse{
		Timestamp: now.Format(time.RFC3339),
		RequestID: requestID,
		Quality:   quality,
	}

	return resp, nil
}
