package handlers

import (
	"html/template"
	"io"
	"net/http"

	"swadesh-dashboard/internal/database"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer wraps the template engine for Echo
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(pattern string) *TemplateRenderer {
	return &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}

// Render implements echo.Renderer interface
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// SetupRoutes configures all HTTP routes
func SetupRoutes(e *echo.Echo, sseHub *SSEHub) {
	// Serve static files
	e.Static("/public", "public")

	// Main page
	e.GET("/", indexHandler)

	// SSE endpoint
	e.GET("/events", sseHub.Handler())

	// HTMX partials
	e.GET("/partials/status-card", statusCardHandler)
	e.GET("/partials/work-order", workOrderHandler)

	// API endpoints
	e.GET("/api/predictions", predictionsHandler)
	e.POST("/api/alerts/:id/dismiss", dismissAlertHandler)
}

// indexHandler renders the main dashboard page
func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

// statusCardHandler renders the AI status card partial
func statusCardHandler(c echo.Context) error {
	label := c.QueryParam("label")
	confidence := c.QueryParam("confidence")

	data := map[string]interface{}{
		"Label":      label,
		"Confidence": confidence,
	}

	return c.Render(http.StatusOK, "status_card.html", data)
}

// workOrderHandler renders the work order modal content
func workOrderHandler(c echo.Context) error {
	label := c.QueryParam("label")
	confidence := c.QueryParam("confidence")

	data := map[string]interface{}{
		"Label":      label,
		"Confidence": confidence,
	}

	return c.Render(http.StatusOK, "work_order_modal.html", data)
}

// predictionsHandler returns recent prediction logs as JSON
func predictionsHandler(c echo.Context) error {
	predictions, err := database.GetRecentPredictions(50)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, predictions)
}
