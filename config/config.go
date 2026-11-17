package config

import (
	"os"
)

type Config struct {
	Port                 string
	MediaServiceURL      string
	MediaServiceGRPCAddr string
	CatalogServiceURL    string
	CatalogServiceGRPCAddr string
	SearchServiceURL     string
	SearchServiceGRPCAddr string
	UserServiceURL       string
	UserServiceGRPCAddr  string
}

func GetEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func LoadConfig() Config {
	return Config{
		Port:                   GetEnv("PORT", "8085"),
		MediaServiceURL:        GetEnv("MEDIA_SERVICE_URL", "http://media-service:8081"),
		MediaServiceGRPCAddr:   GetEnv("MEDIA_SERVICE_GRPC_ADDR", "media-service:50051"),
		CatalogServiceURL:      GetEnv("CATALOG_SERVICE_URL", "http://catalog-service:8082"),
		CatalogServiceGRPCAddr: GetEnv("CATALOG_SERVICE_GRPC_ADDR", "catalog-service:50052"),
		SearchServiceURL:       GetEnv("SEARCH_SERVICE_URL", "http://search-service:8083"),
		SearchServiceGRPCAddr:  GetEnv("SEARCH_SERVICE_GRPC_ADDR", "search-service:50053"),
		UserServiceURL:         GetEnv("USER_SERVICE_URL", "http://user-service:8084"),
		UserServiceGRPCAddr:    GetEnv("USER_SERVICE_GRPC_ADDR", "user-service:50054"),
	}
}
