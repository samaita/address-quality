// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"address-quality/internal/logger"
	"address-quality/internal/model"
)

const (
	googleMapsRequestTimeout = 10 * time.Second
)

var errGoogleMapsNotFound = errors.New("address not found by geocoder")

// canonicalizeRequestBody trims surrounding whitespace and collapses every run
// of whitespace (including newlines) into a single space, producing the
// deterministic single-line body used for the cache request_hash.
func canonicalizeRequestBody(rawBody string) string {
	return strings.Join(strings.Fields(rawBody), " ")
}

func requestHash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// responseIsNotFound reports whether a cached/mocked/real geocoding JSON
// payload represents "no result": an error response, an empty results array,
// or an empty object (the v4 API returns `{}` with HTTP 200 when nothing is
// found).
func responseIsNotFound(payload []byte) bool {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 || string(trimmed) == "{}" {
		return true
	}
	var probe struct {
		Error   *json.RawMessage  `json:"error"`
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return false
	}
	if probe.Error != nil {
		return true
	}
	return probe.Results != nil && len(probe.Results) == 0
}

func mockNotFoundResponse() []byte {
	return []byte(`{"error":{"code":404,"message":"Address not found","status":"NOT_FOUND"}}`)
}

func (svc *Service) ValidateAddressV0(ctx context.Context, rawBody string, requestID string) (*model.GeocodeResponse, int, error) {
	log := logger.L.With().Str("request_id", requestID).Logger()

	canonical := canonicalizeRequestBody(rawBody)
	hash := requestHash(canonical)

	log.Debug().Str("request_hash", hash).Str("request", canonical).Msg("geocode request received")

	var geocodeReq model.GeocodeRequest
	if err := json.Unmarshal([]byte(canonical), &geocodeReq); err != nil {
		return nil, http.StatusBadRequest, errors.Join(ErrValidation, fmt.Errorf("invalid request body: %w", err))
	}

	// The check -> fetch -> cache section is serialized so that concurrent
	// identical requests only trigger a single upstream Google Maps call.
	svc.googleMapsMu.Lock()
	defer svc.googleMapsMu.Unlock()

	if rec, found, err := svc.repo.GetGoogleMapsCache(ctx, hash); err != nil {
		log.Error().Err(err).Str("request_hash", hash).Msg("lookup geocode cache")
		return nil, http.StatusInternalServerError, err
	} else if found {
		log.Debug().Str("request_hash", hash).Msg("geocode cache hit")
		return svc.buildGeocodeResponse(requestID, []byte(rec.Response))
	}

	var payload []byte
	var status int
	var err error
	if svc.googleMapsMock {
		log.Info().Str("request_hash", hash).Msg("geocode cache miss, serving mocked 404")
		payload = mockNotFoundResponse()
		status = http.StatusNotFound
	} else {
		payload, status, err = svc.callGoogleMaps(ctx, geocodeReq.Address)
		if err != nil {
			if errors.Is(err, errGoogleMapsNotFound) {
				status = http.StatusNotFound
			} else {
				log.Error().Err(err).Str("request_hash", hash).Msg("google maps geocode failed")
				return nil, http.StatusInternalServerError, err
			}
		}
	}

	if err := svc.repo.UpsertGoogleMapsCache(ctx, hash, canonical, string(payload), time.Now().UTC()); err != nil {
		log.Error().Err(err).Str("request_hash", hash).Msg("store geocode cache")
		return nil, http.StatusInternalServerError, err
	}
	log.Debug().Str("request_hash", hash).Int("status", status).Msg("geocode response cached")

	return svc.buildGeocodeResponse(requestID, payload)
}

func (svc *Service) callGoogleMaps(ctx context.Context, address string) ([]byte, int, error) {
	endpoint := fmt.Sprintf("%s/%s", strings.TrimSuffix(svc.googleMapsBaseURL, "/"), url.QueryEscape(address))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-Goog-Api-Key", svc.googleMapsAPIKey)

	logger.Info().Str("url", endpoint).Msg("calling google maps geocoding api")

	client := &http.Client{Timeout: googleMapsRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, http.StatusOK, nil
	case http.StatusNotFound:
		return body, http.StatusNotFound, errGoogleMapsNotFound
	default:
		return nil, 0, fmt.Errorf("google maps geocoding returned status %d: %s", resp.StatusCode, string(body))
	}
}

func (svc *Service) buildGeocodeResponse(requestID string, payload []byte) (*model.GeocodeResponse, int, error) {
	status := http.StatusOK
	if responseIsNotFound(payload) {
		status = http.StatusNotFound
	}
	return &model.GeocodeResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
		Data:      json.RawMessage(payload),
	}, status, nil
}
