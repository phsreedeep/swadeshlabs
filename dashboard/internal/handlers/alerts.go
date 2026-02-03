package handlers

import (
	"net/http"
	"swadesh-dashboard/internal/database"
	"github.com/labstack/echo/v4"
)

func DismissAlert(c echo.Context) error {
	id := c.Param("id")
	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	err := database.UpdateDismissReason(id, input.Reason)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Alert dismissed"})
}

func GetAlerts(c echo.Context) error {
	// Implement logic to fetch alerts from database
	return c.JSON(http.StatusOK, []string{}) 
}

func CreateWorkOrder(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{"status": "Work order created"})
}
