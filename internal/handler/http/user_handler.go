package http

import (
	"GoPortfolio/internal/domain"
	"GoPortfolio/internal/service"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(auth *service.AuthService) *UserHandler {
	return &UserHandler{
		authService: auth,
	}
}

func (h *UserHandler) CreateUser(ctx *gin.Context) {
	const op = "http.Handler.CreateUser"
	slog.Info("Request createUser", "method", ctx.Request.Method, "url", ctx.Request.URL.String())

	var user *domain.User
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err = h.authService.RegisterUser(ctx.Request.Context(), user)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
