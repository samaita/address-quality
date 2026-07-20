package model

import (
	"errors"
	"fmt"
)

type AddressRequest struct {
	Address    string `json:"address" validate:"required"`
	SourceCode string `json:"source_code"`
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

type Candidate struct {
	LocationID int64
	Name       string
	Level      string
	Score      float64
	Source     string
	MatchType  string
	PostalCode string
}

type LevelExplain struct {
	Input      string   `json:"input,omitempty"`
	Candidates []string `json:"candidates,omitempty"`
	Winner     string   `json:"winner,omitempty"`
	Reasons    []string `json:"reasons,omitempty"`
}

type Explainability struct {
	Province    *LevelExplain `json:"province,omitempty"`
	City        *LevelExplain `json:"city,omitempty"`
	District    *LevelExplain `json:"district,omitempty"`
	SubDistrict *LevelExplain `json:"subdistrict,omitempty"`
}

type Quality struct {
	AddressID       string         `json:"address_id"`
	Confidence      float64        `json:"confidence"`
	Location        Location       `json:"location"`
	NormalizedInput string         `json:"normalized_input"`
	FormattedOutput string         `json:"formatted_output"`
	LocationVersion string         `json:"location_version"`
	LocationSource  string         `json:"location_source"`
	RawInput        string         `json:"raw_input"`
	Explainability  Explainability `json:"-"`
}

type AddressResponse struct {
	Timestamp string  `json:"timestamp"`
	RequestID string  `json:"request_id"`
	Quality   Quality `json:"quality"`
}
