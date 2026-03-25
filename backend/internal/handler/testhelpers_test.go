package handler

import (
	"time"

	"github.com/newa/zeiterfassung/internal/config"
)

var testConfig = config.Config{
	Port: 8080,
	Env:  "development",
	FrontendURL: "http://localhost:3000",
	SessionSecret: "test-secret",
	SessionMaxAge: 24 * time.Hour,
	AzureAD: config.AzureADConfig{
		TenantID:     "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost:8080/api/auth/callback",
	},
}
