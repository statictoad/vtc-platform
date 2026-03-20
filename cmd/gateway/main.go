package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := getEnv("PORT", "8080")
	slog.Info("gateway starting", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
