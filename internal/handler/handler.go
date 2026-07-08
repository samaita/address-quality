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

func errorResponse(c echo.Context, status int, msg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	requestID := uuid.Must(uuid.NewV7()).String()
	resp := map[string]string{
		"timestamp":  now,
		"request_id": requestID,
		"error":      msg,
	}
	return c.JSON(status, resp)
}

func HandleAddressRequest(c echo.Context) error {
	var req model.AddressRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if req.Address == "" {
		return errorResponse(c, http.StatusBadRequest, "address is required")
	}

	now := time.Now().UTC()
	requestID := uuid.Must(uuid.NewV7()).String()
	addressID := uuid.Must(uuid.NewV7()).String()

	rawInput := req.Address
	sanitizedAddr := sanitizer.Sanitize(req.Address)

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

	outputJSON, _ := json.Marshal(sanitizedAddr)

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

	if err := database.InsertRecord(record); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "failed to store record")
	}

	return c.JSON(http.StatusOK, resp)
}
