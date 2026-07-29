// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package model

import (
	"errors"
	"fmt"
)

type HealthResponse struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database,omitempty" example:"ok"`
}

type ErrorResponse struct {
	Timestamp string `json:"timestamp" example:"2026-07-29T05:06:43Z"`
	RequestID string `json:"request_id" example:"019fac44-95c0-79cb-b1d4-649463403ea7"`
	Error     string `json:"error" example:"invalid request body"`
}

type AddressRequest struct {
	Address    string `json:"address" validate:"required" example:"JL. Supratman No.72, Citarum, 40191"`
	SourceCode string `json:"source_code" example:"kemendagri"`
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
	PostalCode  string `json:"postal_code" example:"40115"`
	SubDistrict string `json:"sub_district" example:"Citarum"`
	District    string `json:"district" example:"Bandung Wetan"`
	City        string `json:"city" example:"Kota Bandung"`
	Province    string `json:"province" example:"Jawa Barat"`
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
	Type    string `json:"type" example:"postal_code_mismatch"`
	Message string `json:"message" example:"Postal code 40115 does not match district Citarum"`
}

type Reason string

const (
	ReasonExactMatch         Reason = "exact_match"
	ReasonMatchProvince      Reason = "match_province"
	ReasonMatchCity          Reason = "match_city"
	ReasonMatchDistrict      Reason = "match_district"
	ReasonMatchSubDistrict   Reason = "match_subdistrict"
	ReasonPostalCodeExact    Reason = "match_postal_code_exact"
	ReasonPostalCodePrefix4  Reason = "match_postal_code_prefix4"
	ReasonPostalCodePrefix3  Reason = "match_postal_code_prefix3"
	ReasonPostalCodeLookup   Reason = "postal_code_lookup"
	ReasonPostalCodeInferred Reason = "postal_code_inferred"
)

type CandidateEvaluation struct {
	Candidate     AdminCandidate
	Confidence    float64
	Status        QualityStatus
	Matched       []Component
	Missing       []Component
	UnusedEvidence []Evidence
	Conflicts     []Conflict
	Reasons       []Reason
}

type Assessment struct {
	Matched   []string   `json:"matched"`
	Missing   []string   `json:"missing"`
	Conflicts []Conflict `json:"conflicts"`
	Ambiguous []string   `json:"ambiguous"`
}

type Resolution struct {
	Strategy       []string              `json:"strategy"`
	CandidateCount int                   `json:"candidate_count" example:"1"`
	Candidates     []ResolutionCandidate `json:"candidates"`
}

type ResolutionCandidate struct {
	UUID     string   `json:"uuid" example:"019fac45-d6d0-7e53-8e0d-a44f30d72a53"`
	Score    float64  `json:"score" example:"0.35"`
	Location Location `json:"location"`
	Reasons  []string `json:"reasons"`
}

type Metadata struct {
	LocationSource  string `json:"location_source" example:"kemendagri"`
	LocationVersion string `json:"location_version" example:"2025"`
}

type ResponseData struct {
	AddressID       string        `json:"address_id" example:"019fac45-d6cb-7153-aeef-742c66db6d18"`
	Status          QualityStatus `json:"status" example:"VALID"`
	Confidence      float64       `json:"confidence" example:"0.35"`
	RawInput        string        `json:"raw_input" example:"JL. Supratman No.72, Citarum, 40191"`
	NormalizedInput string        `json:"normalized_input" example:"jl supratman no citarum 40191"`
	FormattedAddr   string        `json:"formatted_address" example:"Citarum, Bandung Wetan, Kota Bandung, Jawa Barat 40115"`
	Location        Location      `json:"location"`
	Assessment      Assessment    `json:"assessment"`
	Resolution      Resolution    `json:"resolution"`
	Metadata        Metadata      `json:"metadata"`
}

type AddressResponse struct {
	Timestamp string       `json:"timestamp" example:"2026-07-29T05:08:05Z"`
	RequestID string       `json:"request_id" example:"019fac45-d6cb-7101-9159-76bd7c25867b"`
	Data      ResponseData `json:"data"`
}

type EvidenceType string

const (
	EvidenceRoadName   EvidenceType = "road_name"
	EvidencePlaceName  EvidenceType = "place_name"
	EvidencePostalCode EvidenceType = "postal_code"
)

type Evidence struct {
	Type  EvidenceType
	Value string
}

type Entity struct {
	ID         int64
	Name       string
	Level      string
	PostalCode string
}

type ResolvedEvidence struct {
	Evidence
	Candidates []Entity
}

type DiscoveryStrategy string

const (
	DiscoveryTopDown  DiscoveryStrategy = "top_down"
	DiscoveryAnyLevel DiscoveryStrategy = "any_level"
	DiscoveryAlias    DiscoveryStrategy = "alias"
	DiscoveryPostal   DiscoveryStrategy = "postal"
)

type MatchedEvidence struct {
	Evidence
	Resolved *Entity
}

type Province struct {
	ID             int64
	Name           string
	NormalizedName string
}

type City struct {
	ID             int64
	Name           string
	NormalizedName string
	PostalCode     string
}

type District struct {
	ID             int64
	Name           string
	NormalizedName string
}

type SubDistrict struct {
	ID             int64
	Name           string
	NormalizedName string
	PostalCode     string
}

type PostalCode struct {
	ID    int64
	Code  string
	Name  string
}

type Road struct {
	ID   int64
	Name string
}

type AdminLocation struct {
	Province    *Province
	City        *City
	District    *District
	SubDistrict *SubDistrict
	PostalCode  *PostalCode
	Road        *Road
	Conflicts   []Conflict
}

type CandidateOriginLevel string

const (
	OriginProvince    CandidateOriginLevel = "PROVINCE"
	OriginCity        CandidateOriginLevel = "CITY"
	OriginDistrict    CandidateOriginLevel = "DISTRICT"
	OriginSubDistrict CandidateOriginLevel = "SUBDISTRICT"
)

type AdminCandidate struct {
	UUID                string
	OriginLevel         CandidateOriginLevel
	Location            AdminLocation
	Evidence            []MatchedEvidence
	DiscoveryStrategies []DiscoveryStrategy
}
