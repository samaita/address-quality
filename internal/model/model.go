// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

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

type QualityStatus string

const (
	StatusValid      QualityStatus = "VALID"
	StatusIncomplete QualityStatus = "INCOMPLETE"
	StatusAmbiguous  QualityStatus = "AMBIGUOUS"
	StatusConflict   QualityStatus = "CONFLICT"
	StatusUnknown    QualityStatus = "UNKNOWN"
)

type Component string

const (
	ComponentProvince    Component = "province"
	ComponentCity        Component = "city"
	ComponentDistrict    Component = "district"
	ComponentSubDistrict Component = "sub_district"
	ComponentPostalCode  Component = "postal_code"
)

type Conflict struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type CandidateEvaluation struct {
	Confidence float64
	Status     QualityStatus
	Matched    []Component
	Missing    []Component
	Conflicts  []Conflict
	Reasons    []string
}

type Assessment struct {
	Matched   []string   `json:"matched"`
	Missing   []string   `json:"missing"`
	Conflicts []Conflict `json:"conflicts"`
	Ambiguous []string   `json:"ambiguous"`
}

type Resolution struct {
	Strategy       []string             `json:"strategy"`
	CandidateCount int                  `json:"candidate_count"`
	Candidates     []ResolutionCandidate `json:"candidates"`
}

type ResolutionCandidate struct {
	Score    float64  `json:"score"`
	Location Location `json:"location"`
	Reasons  []string `json:"reasons"`
}

type Metadata struct {
	LocationSource  string `json:"location_source"`
	LocationVersion string `json:"location_version"`
}

type ResponseData struct {
	AddressID       string        `json:"address_id"`
	Status          QualityStatus `json:"status"`
	Confidence      float64       `json:"confidence"`
	RawInput        string        `json:"raw_input"`
	NormalizedInput string        `json:"normalized_input"`
	FormattedAddr   string        `json:"formatted_address"`
	Location        Location      `json:"location"`
	Assessment      Assessment    `json:"assessment"`
	Resolution      Resolution    `json:"resolution"`
	Metadata        Metadata      `json:"metadata"`
}

type AddressResponse struct {
	Timestamp string       `json:"timestamp"`
	RequestID string       `json:"request_id"`
	Data      ResponseData `json:"data"`
}
