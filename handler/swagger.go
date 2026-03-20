package handler

import (
	"net/http"

	"github.com/EdgeFlowCDN/cdn-control/docs"
	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>EdgeFlow API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/swagger/doc.json",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    });
  </script>
</body>
</html>`

// RegisterSwaggerRoutes adds the Swagger UI and spec routes to the router.
func RegisterSwaggerRoutes(r *gin.Engine) {
	r.GET("/swagger/doc.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(docs.OpenAPISpec))
	})

	r.GET("/swagger/*any", func(c *gin.Context) {
		// Serve the Swagger UI HTML for any sub-path under /swagger/
		// (except doc.json which is handled above with a more specific route).
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
}
