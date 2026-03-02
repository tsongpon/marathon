package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/tsongpon/marathon/internal/model"
)

type createAlertRequest struct {
	Title    string `json:"title"`
	Details  string `json:"details"`
	Severity string `json:"severity"`
}

type AlertHttpHandler struct {
	service AlertService
}

func NewAlertHttpHandler(service AlertService) *AlertHttpHandler {
	return &AlertHttpHandler{service: service}
}

func (h *AlertHttpHandler) CreateGenericAlerts(c *echo.Context) error {
	var req createAlertRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
	}

	alert := model.Alert{
		Title:    req.Title,
		Details:  req.Details,
		Severity: req.Severity,
	}

	created, err := h.service.CreateAlert(c.Request().Context(), alert)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create alert"})
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *AlertHttpHandler) CreateSignozAlert(c *echo.Context) error {

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Print raw JSON body
	println(string(body))
	return c.String(http.StatusOK, "ok")
}

func (h *AlertHttpHandler) NotifyAlerts(c *echo.Context) error {
	go h.service.Notify(context.Background(), time.Now())

	return c.JSON(http.StatusAccepted, map[string]string{"status": "notification job started"})
}

func (h *AlertHttpHandler) DeleteAlerts(c *echo.Context) error {
	return nil
}
