package middleware

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func SwaggerHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)
}

func SwaggerRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

// SwaggerConfig holds swagger configuration
type SwaggerConfig struct {
	Enabled  bool
	BasePath string
	Host     string
}

func NewSwaggerConfig(enabled bool, host, basePath string) *SwaggerConfig {
	return &SwaggerConfig{
		Enabled:  enabled,
		Host:     host,
		BasePath: basePath,
	}
}
