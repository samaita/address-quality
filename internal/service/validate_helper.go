package service

import (
	"time"

	"encoding/json"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

func (svc *Service) sanitize(input string) string {
	return svc.s.Sanitize(input)
}

func buildAddressRecord(requestID, addressID string, quality model.Quality, now time.Time) *database.AddressRecord {
	outputJSON, _ := json.Marshal(quality)

	return &database.AddressRecord{
		ID:              requestID,
		AddressID:       addressID,
		RawInput:        quality.RawInput,
		NormalizedAddr:  quality.Output,
		Confidence:      quality.Confidence,
		PostalCode:      quality.Location.PostalCode,
		SubDistrict:     quality.Location.SubDistrict,
		District:        quality.Location.District,
		City:            quality.Location.City,
		Province:        quality.Location.Province,
		LocationVersion: quality.LocationVersion,
		OutputJSON:      string(outputJSON),
		CreatedAt:       now,
	}
}
