package main

import (
	"GoPortfolio/internal/configLoader"
	"GoPortfolio/internal/repository/mysql"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	prepareEnv()
	cfg := configLoader.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := sql.Open("mysql", "./db/mysql.db")
	if err != nil {
		slog.Error("MySQL connection failed")
	}
	repo := mysql.NewMysqlUserRepo(db)
	_ = repo
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/", testHandler)
	srv := http.Server{
		Addr:         cfg.HttpSrv.Address,
		Handler:      router,
		IdleTimeout:  cfg.HttpSrv.IdleTimeout,
		ReadTimeout:  cfg.HttpSrv.Timeout,
		WriteTimeout: cfg.HttpSrv.Timeout,
	}
	slog.Info("Server starting on " + cfg.HttpSrv.Address)
	err := srv.ListenAndServe()
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

func testHandler(ctx *gin.Context) {
	slog.Info("Request received", "method", ctx.Request.Method, "url", ctx.Request.URL.String())
	ctx.String(http.StatusOK, "Hello, Go!")
}
