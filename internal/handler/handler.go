package handler

import (
	"net/http"
	"time"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/sanitizer"
)

type Handler struct {
	repo             *database.Repository
	s                *sanitizer.Sanitizer
	maxAddressLength int
}

func New(repo *database.Repository, s *sanitizer.Sanitizer, maxAddressLength int) *Handler {
	return &Handler{repo: repo, s: s, maxAddressLength: maxAddressLength}
}

func (h *Handler) HandleHealthCheck(c echo.Context) error {
	ctx := c.Request().Context()
	if err := h.repo.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func errorResponse(c echo.Context, status int, msg string, requestID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	resp := map[string]string{
		"timestamp":  now,
		"request_id": requestID,
		"error":      msg,
	}
	return c.JSON(status, resp)
}

func (h *Handler) HandleAddressRequest(c echo.Context) error {
	requestID := uuid.Must(uuid.NewV7()).String()

	var req model.AddressRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "invalid request body", requestID)
	}

	if req.Address == "" {
		return errorResponse(c, http.StatusBadRequest, "address is required", requestID)
	}

	if len(req.Address) > h.maxAddressLength {
		return errorResponse(c, http.StatusBadRequest, "address exceeds maximum length of 1000 characters", requestID)
	}

	now := time.Now().UTC()
	addressID := uuid.Must(uuid.NewV7()).String()

	rawInput := req.Address
	sanitizedAddr := h.s.Sanitize(req.Address)

	quality := model.Quality{
		AddressID:       addressID,
		Confidence:      0.0,
		Location:        model.Location{},
		NormalizedInput: sanitizedAddr,
		Output:          sanitizedAddr,
		LocationVersion: "",
		RawInput:        rawInput,
	}

	resp := model.AddressResponse{
		Timestamp: now.Format(time.RFC3339),
		RequestID: requestID,
		Quality:   quality,
	}

	outputJSON, _ := json.Marshal(quality)

	record := &database.AddressRecord{
		ID:               requestID,
		AddressID:        addressID,
		RawInput:         rawInput,
		NormalizedAddr:   sanitizedAddr,
		Confidence:       0.0,
		PostalCode:       "",
		SubDistrict:      "",
		District:         "",
		City:             "",
		Province:         "",
		LocationVersion:  "",
		OutputJSON:       string(outputJSON),
		CreatedAt:        now,
	}

	ctx := c.Request().Context()
	if err := h.repo.InsertRecord(ctx, record); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "failed to store record", requestID)
	}

	return c.JSON(http.StatusOK, resp)
}
