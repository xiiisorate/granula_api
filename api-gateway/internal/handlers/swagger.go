// =============================================================================
// Package handlers provides HTTP request handlers for API Gateway.
// =============================================================================
// This file contains Swagger UI handler for serving API documentation.
// =============================================================================
package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SwaggerHandler serves Swagger UI for API documentation.
type SwaggerHandler struct {
	specPath string
}

// NewSwaggerHandler creates a new SwaggerHandler.
func NewSwaggerHandler(specPath string) *SwaggerHandler {
	return &SwaggerHandler{specPath: specPath}
}

// ServeUI returns an HTML page with Swagger UI.
// @Summary Swagger UI
// @Description Interactive API documentation
// @Tags documentation
// @Produce html
// @Router /swagger [get]
func (h *SwaggerHandler) ServeUI(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Granula API - Swagger Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://granula.ru/favicon.png">
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1a1a2e;
            padding: 10px 0;
        }
        .swagger-ui .topbar .download-url-wrapper .download-url-button {
            background-color: #e94560;
        }
        .swagger-ui .topbar .download-url-wrapper .download-url-button:hover {
            background-color: #ff6b6b;
        }
        .swagger-ui .info .title {
            color: #1a1a2e;
        }
        .swagger-ui .info .title small {
            background-color: #e94560;
        }
        .swagger-ui .opblock.opblock-post {
            border-color: #49cc90;
            background: rgba(73, 204, 144, .1);
        }
        .swagger-ui .opblock.opblock-get {
            border-color: #61affe;
            background: rgba(97, 175, 254, .1);
        }
        .swagger-ui .opblock.opblock-put {
            border-color: #fca130;
            background: rgba(252, 161, 48, .1);
        }
        .swagger-ui .opblock.opblock-delete {
            border-color: #f93e3e;
            background: rgba(249, 62, 62, .1);
        }
        .swagger-ui .btn.execute {
            background-color: #e94560;
            border-color: #e94560;
        }
        .swagger-ui .btn.execute:hover {
            background-color: #ff6b6b;
            border-color: #ff6b6b;
        }
        .topbar-wrapper img {
            content: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 30'%3E%3Ctext x='5' y='22' font-family='Arial, sans-serif' font-size='20' font-weight='bold' fill='%23ffffff'%3EGranula API%3C/text%3E%3C/svg%3E");
        }
        .swagger-ui .scheme-container {
            background: #f7f7f7;
            box-shadow: 0 1px 2px 0 rgba(0,0,0,.15);
        }
        /* Custom header */
        .custom-header {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: white;
            padding: 20px;
            text-align: center;
        }
        .custom-header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 700;
        }
        .custom-header p {
            margin: 10px 0 0;
            opacity: 0.9;
            font-size: 14px;
        }
        .custom-header .version {
            display: inline-block;
            background: #e94560;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="custom-header">
        <h1>🏠 Granula API</h1>
        <p>Интеллектуальный сервис планирования ремонта и перепланировки</p>
        <span class="version">v1.0.0</span>
    </div>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger/spec.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
                tryItOutEnabled: true,
                requestSnippetsEnabled: true,
                persistAuthorization: true,
                displayRequestDuration: true,
                syntaxHighlight: {
                    activate: true,
                    theme: "monokai"
                },
                defaultModelsExpandDepth: 2,
                defaultModelExpandDepth: 2
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// ServeSpec returns the OpenAPI specification file.
// @Summary OpenAPI Specification
// @Description Returns the OpenAPI 3.0 specification in YAML format
// @Tags documentation
// @Produce text/yaml
// @Router /swagger/spec.yaml [get]
func (h *SwaggerHandler) ServeSpec(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/yaml; charset=utf-8")
	c.Set("Access-Control-Allow-Origin", "*")
	// Prevent caching to always serve fresh spec
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	return c.SendFile(h.specPath)
}

// ServeSpecJSON returns the OpenAPI specification in JSON format.
// @Summary OpenAPI Specification (JSON)
// @Description Returns the OpenAPI 3.0 specification in JSON format
// @Tags documentation
// @Produce application/json
// @Router /swagger/spec.json [get]
func (h *SwaggerHandler) ServeSpecJSON(c *fiber.Ctx) error {
	// For now, redirect to YAML - JSON conversion can be added later
	return c.Redirect("/swagger/spec.yaml", fiber.StatusTemporaryRedirect)
}

