// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"address-quality/internal/database"
	"address-quality/internal/logger"
	"address-quality/internal/model"
	"address-quality/internal/queue"
	"address-quality/internal/sanitizer"
)

var ErrValidation = errors.New("validation error")

const (
	storeQueueSize     = 1000
	storeQueueWorkers  = 2
	storeQueueJobLimit = 5 * time.Second
)

type AddressRepository interface {
	InsertAddressRequest(ctx context.Context, rec *database.AddressRecord) error
	Ping(ctx context.Context) error
	GetGoogleMapsCache(ctx context.Context, requestHash string) (*database.GoogleMapsCacheRecord, bool, error)
	UpsertGoogleMapsCache(ctx context.Context, requestHash, request, response string, now time.Time) error
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
	FindAllCityPriority(ctx context.Context, sourceID int64) ([]database.CityPriorityRow, error)
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
	repo               AddressRepository
	locationRepo       LocationRepository
	s                  *sanitizer.Sanitizer
	maxAddressLength   int
	sourceCode         string
	enableStoreRequest bool
	storeQueue         *queue.Queue[*database.AddressRecord]

	googleMapsMock   bool
	googleMapsAPIKey string
	googleMapsBaseURL string
	googleMapsMu     sync.Mutex
	provinceCache       map[string][]*provinceEntry
	provinceOnce        sync.Once
	provinceErr         error
	provinceKodeToEntry map[string]*provinceEntry
	provinceByID        map[int64]*provinceEntry

	cityCache map[string][]*cityEntry
	cityOnce  sync.Once
	cityErr   error
	cityByID  map[int64]*cityEntry

	districtCache map[string][]*districtEntry
	districtOnce  sync.Once
	districtErr   error
	districtByID  map[int64]*districtEntry

	subDistrictCache map[string][]*subDistrictEntry
	subDistrictOnce  sync.Once
	subDistrictErr   error
	subDistrictByID  map[int64]*subDistrictEntry

	cityPrioritySet  map[string]string // normalized name -> "KOTA" | "KABUPATEN"
	cityPriorityOnce sync.Once
	cityPriorityErr  error

	hierarchyCache *database.HierarchyMap
	hierarchyOnce  sync.Once
	hierarchyErr   error

	phraseDict     map[string]map[string][]model.Entity
	phraseDictOnce sync.Once
	phraseDictErr  error
}

func New(repo AddressRepository, locationRepo LocationRepository, s *sanitizer.Sanitizer, maxAddressLength int, sourceCode string, enableStoreRequest bool, googleMapsMock bool, googleMapsAPIKey string, googleMapsBaseURL string) *Service {
	svc := &Service{
		repo:               repo,
		locationRepo:       locationRepo,
		s:                  s,
		maxAddressLength:   maxAddressLength,
		sourceCode:         sourceCode,
		enableStoreRequest: enableStoreRequest,
		googleMapsMock:     googleMapsMock,
		googleMapsAPIKey:   googleMapsAPIKey,
		googleMapsBaseURL:  googleMapsBaseURL,
	}

	if enableStoreRequest {
		svc.storeQueue = queue.New(storeQueueSize, storeQueueWorkers, func(ctx context.Context, rec *database.AddressRecord) {
			jobCtx, cancel := context.WithTimeout(ctx, storeQueueJobLimit)
			defer cancel()
			if err := svc.repo.InsertAddressRequest(jobCtx, rec); err != nil {
				logger.Warn().Err(err).Str("request_id", rec.ID).Msg("failed to store address request")
			}
		})
	}

	return svc
}

// Shutdown stops the store queue from accepting new jobs and waits for all
// queued address records to be written before returning. It is a no-op when
// request storage is disabled.
func (svc *Service) Shutdown(ctx context.Context) error {
	if svc.storeQueue == nil {
		return nil
	}
	return svc.storeQueue.Shutdown(ctx)
}

func (svc *Service) Ping(ctx context.Context) error {
	return svc.repo.Ping(ctx)
}

func (svc *Service) MaxAddressLength() int {
	return svc.maxAddressLength
}

func (svc *Service) ValidateAddress(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	return svc.ValidateAddressV1(ctx, req, requestID)
}

// ValidateGeocodeAddress resolves an address through the Google Maps geocoding
// API using an sqlite lazy cache. It takes the raw request body so the cache
// key is derived from the trimmed, single-line body.
func (svc *Service) ValidateGeocodeAddress(ctx context.Context, rawBody string, requestID string) (*model.GeocodeResponse, int, error) {
	return svc.ValidateAddressV0(ctx, rawBody, requestID)
}
