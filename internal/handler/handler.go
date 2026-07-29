// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	mw "address-quality/internal/middleware"
	"address-quality/internal/model"
	"address-quality/internal/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// HandleHealthCheck checks database connectivity.
// @Summary      Health check
// @Description  Checks database connectivity and returns service status
// @Success      200  {object}  model.HealthResponse
// @Failure      503  {object}  model.HealthResponse
// @Router       /health [get]
func (h *Handler) HandleHealthCheck(c echo.Context) error {
	ctx := c.Request().Context()
	if err := h.svc.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, model.HealthResponse{Status: "error"})
	}
	return c.JSON(http.StatusOK, model.HealthResponse{Status: "ok", Database: "ok"})
}

func errorResponse(c echo.Context, status int, msg string, requestID string) error {
	resp := model.ErrorResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
		Error:     msg,
	}
	return c.JSON(status, resp)
}

// HandleAddressRequest validates and resolves an address.
// @Summary      Validate an address
// @Description  Validate and resolve a Thai address against the official location database
// @Accept       json
// @Produce      json
// @Param        request  body  model.AddressRequest  true  "Address to validate"
// @Success      200  {object}  model.AddressResponse
// @Failure      400  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /v1/validate [post]
func (h *Handler) HandleAddressRequest(c echo.Context) error {
	requestID := mw.GetRequestID(c.Request().Context())

	var req model.AddressRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "invalid request body", requestID)
	}

	resp, err := h.svc.ValidateAddress(c.Request().Context(), &req, requestID)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return errorResponse(c, http.StatusBadRequest, err.Error(), requestID)
		}
		return errorResponse(c, http.StatusInternalServerError, "failed to store record", requestID)
	}

	return c.JSON(http.StatusOK, resp)
}
