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

func (h *Handler) HandleHealthCheck(c echo.Context) error {
	ctx := c.Request().Context()
	if err := h.svc.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "error"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
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
