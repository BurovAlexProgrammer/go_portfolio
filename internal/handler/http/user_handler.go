package http

import (
	"GoPortfolio/internal/domain"
	"GoPortfolio/internal/usecase"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: uc,
	}
}

func (h *UserHandler) CreateUser(ctx *gin.Context) {
	const op = "http.Handler.CreateUser"
	slog.Info("Request createUser", "method", ctx.Request.Method, "url", ctx.Request.URL.String())

	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUsecase.CreateUser(ctx.Request.Context(), &user); err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
