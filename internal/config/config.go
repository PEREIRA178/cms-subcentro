package config

import (
	"os"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port               string
	Env                string
	CORSOrigins        string
	StaticCacheDuration time.Duration

	// PocketBase
	PBUrl    string
	PBAdmin  string
	PBPass   string

	// JWT
	JWTSecret     string
	JWTExpiration time.Duration

	// Cloudflare R2
	R2AccountID   string
	R2AccessKey   string
	R2SecretKey   string
	R2BucketName  string
	R2Region      string
	R2PublicURL   string

	// WhatsApp (Twilio)
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Ollama
	OllamaURL   string
	OllamaModel string

	// Web
	BaseURL     string
	SiteName    string
}

func Load() *Config {
	return &Config{
		// Server
		Port:               getEnv("PORT", "3000"),
		Env:                getEnv("ENV", "development"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
		StaticCacheDuration: 24 * time.Hour,

		// PocketBase
		PBUrl:   getEnv("PB_URL", "http://127.0.0.1:8090"),
		PBAdmin: getEnv("PB_ADMIN_EMAIL", "admin@jcp.cl"),
		PBPass:  getEnv("PB_ADMIN_PASSWORD", ""),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "jcp-secret-change-me-in-production"),
		JWTExpiration: 72 * time.Hour,

		// Cloudflare R2
		R2AccountID:  getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKey:  getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName: getEnv("R2_BUCKET_NAME", "jcp-media"),
		R2Region:     getEnv("R2_REGION", "auto"),
		R2PublicURL:  getEnv("R2_PUBLIC_URL", ""),

		// WhatsApp
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),

		// Ollama
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "llama3"),

		// Web
		BaseURL:  getEnv("BASE_URL", "http://localhost:3000"),
		SiteName: "JCP Gestión Inmobiliaria",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
