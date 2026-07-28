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
	CandidateCount int                   `json:"candidate_count"`
	Candidates     []ResolutionCandidate `json:"candidates"`
}

type ResolutionCandidate struct {
	UUID     string   `json:"uuid"`
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
