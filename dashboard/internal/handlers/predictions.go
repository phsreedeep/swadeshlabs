package handlers

import (
	"net/http"
	"swadesh-dashboard/internal/database"
	"github.com/labstack/echo/v4"
)

func GetPredictions(c echo.Context) error {
	logs, err := database.GetRecentPredictions(50)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, logs)
}

func GetStats(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "active"})
}

func SSEHandler(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	return nil // Requires logic for streaming data
}
