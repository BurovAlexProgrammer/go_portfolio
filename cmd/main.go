package main

import (
	"GoPortfolio/internal/configLoader"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	prepareEnv()
	configLoader.New()

	http.Handle("/", testHandler())

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Server starting on port 8080")
	err := http.ListenAndServe(":8080", testHandler())
	if err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
	slog.Info("Server started on port 8080")
	defer slog.Info("Server stopped")
	defer os.Exit(0)
}

func prepareEnv() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Env file not found")
	}
	fmt.Println("ENV:" + os.Getenv("ENV"))

}

func testHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, Go!")

		if err != nil {
			slog.Error("Error writing response", "error", err)
		}

		slog.Info("Request received", "method", r.Method, "url", r.URL.String())
	}
}
