package main

import (
	"GoPortfolio/internal/configLoader"
	httpHandler "GoPortfolio/internal/handler/http"
	"GoPortfolio/internal/handler/telegram"
	"GoPortfolio/internal/repository/gorm"
	"GoPortfolio/internal/service"
	"GoPortfolio/pkg/orm"
	"fmt"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"sync"

	_ "modernc.org/sqlite"
)

func main() {
	prepareEnv()
	cfg := configLoader.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := ORM.NewPostgresDB(ORM.GetPostgresDSNFromEnv(), cfg.Logger.DbOperationsEnabled)
	if err != nil {
		slog.Error("error: ", err)
		os.Exit(1)
	}

	err = ORM.Automigrate(db)
	if err != nil {
		slog.Error("error: ", err)
		os.Exit(1)
	}

	//Repository
	userGormRepo := gorm.NewUserGormRepo(db)
	taskRepo := gorm.NewTaskGormRepo(db)

	//Services
	authService := service.NewAuthService(userGormRepo)
	taskService := service.NewTaskService(taskRepo)

	//Handler
	httpUserHandler := httpHandler.NewUserHandler(authService)

	//HttpRouter
	router := newRouter(httpUserHandler)

	var wg sync.WaitGroup
	serverStarted := make(chan struct{})
	wg.Add(1)

	go func() {
		startServer(cfg, router, err, serverStarted)
		wg.Done()
	}()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
	}
	bot.Debug = true

	telegramUserHandler := telegram.NewUpdatesHandler(bot, authService, taskService)
	telegramUserHandler.StartUpdates(&wg)

	<-serverStarted
	slog.Info("Server started and working..")
	wg.Wait()
	defer slog.Info("Server stopped")
	defer os.Exit(0)
}

func startServer(cfg *configLoader.AppConfig, router *gin.Engine, err error, startedChan chan struct{}) {
	srv := http.Server{
		Addr:         cfg.HttpSrv.Address,
		Handler:      router,
		IdleTimeout:  cfg.HttpSrv.IdleTimeout,
		ReadTimeout:  cfg.HttpSrv.Timeout,
		WriteTimeout: cfg.HttpSrv.Timeout,
	}
	slog.Info("Server starting on " + cfg.HttpSrv.Address)
	close(startedChan)
	err = srv.ListenAndServe()
	if err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

func newRouter(h *httpHandler.UserHandler) *gin.Engine {
	ginMode := os.Getenv(gin.EnvGinMode)
	gin.SetMode(ginMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	//router.GET("/", testHandler)
	router.POST("/createUser", h.CreateUser)
	return router
}

func prepareEnv() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Env file not found")
	}
	fmt.Println("ENV:" + os.Getenv("ENV"))
}
