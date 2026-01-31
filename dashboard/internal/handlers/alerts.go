package handlers

import (
	"net/http"
	"swadesh-dashboard/internal/database"

	"github.com/labstack/echo/v4"
)

// DismissRequest represents the request body for dismissing an alert
type DismissRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// dismissAlertHandler handles POST /api/alerts/:id/dismiss
func dismissAlertHandler(c echo.Context) error {
	id := c.Param("id")

	var req DismissRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Dismiss reason is required",
		})
	}

	// Update the prediction log with dismiss reason
	if err := database.UpdateDismissReason(id, req.Reason); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "dismissed",
		"message": "Alert dismissed successfully",
	})
}
