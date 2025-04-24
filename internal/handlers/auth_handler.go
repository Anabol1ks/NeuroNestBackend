package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Nickname string `json:"nickname" binding:"required" example:"user123"`
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"yi29jksA"`
}

func RegisterHandler(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Ошибка валиадции",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	var existingUser models.User

	if err := db.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Пользователь с таким email уже существует",
			Code:    "USER_ALREADY_EXISTS",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "PASSWORD_HASH_ERROR",
			Message: "Ошибка при хешировании пароля",
		})
		return
	}

	user := models.User{
		Nickname:     input.Nickname,
		Email:        input.Email,
		PasswordHASH: string(hashedPassword),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "DB_ERROR",
			Message: "Ошибка при создании пользователя",
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Пользователь успешно зарегистрирован",
	})
}
