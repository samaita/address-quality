package model

import (
	"errors"
	"fmt"
)

type AddressRequest struct {
	Address string `json:"address" validate:"required"`
}

func (r *AddressRequest) Validate(maxLength int) error {
	if r.Address == "" {
		return errors.New("address is required")
	}
	if len(r.Address) > maxLength {
		return fmt.Errorf("address exceeds maximum length of %d characters", maxLength)
	}
	return nil
}

type Location struct {
	PostalCode  string `json:"postal_code"`
	SubDistrict string `json:"sub_district"`
	District    string `json:"district"`
	City        string `json:"city"`
	Province    string `json:"province"`
}

type Quality struct {
	AddressID       string   `json:"address_id"`
	Confidence      float64  `json:"confidence"`
	Location        Location `json:"location"`
	NormalizedInput string   `json:"normalized_input"`
	Output          string   `json:"output"`
	LocationVersion string   `json:"location_version"`
	RawInput        string   `json:"raw_input"`
}

type AddressResponse struct {
	Timestamp string  `json:"timestamp"`
	RequestID string  `json:"request_id"`
	Quality   Quality `json:"quality"`
}
