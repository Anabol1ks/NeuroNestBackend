package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	AccessSecret  = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	refreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
)

type RegisterInput struct {
	Nickname string `json:"nickname" binding:"required" example:"user123"`
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"yi29jksA"`
}

// @Summary		Регистрация пользователя
// @Description	Регистрация нового пользователя
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		RegisterInput				true	"Данные пользователя"
// @Success		201		{object}	response.SuccessResponse	"Пользователь успешно зарегистрирован"
// @Failure		400		{object}	response.ErrorResponse		"Ошибка валидации (VALIDATION_ERROR) или пользователь уже существует (EMAIL_EXISTS)"
// @Failure		500		{object}	response.ErrorResponse		"Ошибка сервера (PASSWORD_HASH_ERROR, DB_ERROR)"
// @Router			/auth/register [post]
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
			Code:    "EMAIL_EXISTS",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "PASSWORD_HASH_ERROR",
			Message: "Ошибка сервера",
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

type LoginInput struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"yi29jksA"`
}

// @Summary		Авторизация пользователя
// @Description	Авторизация пользователя и получение токенов
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		LoginInput			true	"Данные для авторизации"
// @Success		200		{object}	response.TokenResponse	"Успешная авторизация"
// @Failure		400		{object}	response.ErrorResponse	"Ошибка валидации данных (VALIDATION_ERROR)"
// @Failure		401		{object}	response.ErrorResponse	"Неверный email или пароль (INVALID_CREDENTIALS)"
// @Failure		500		{object}	response.ErrorResponse	"Ошибка сервера (TOKEN_GENERATION_ERROR)"
// @Router			/auth/login [post]
func LoginHandler(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Ошибка валидации данных",
			Details: err.Error(),
		})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "INVALID_CREDENTIALS",
			Message: "Неверный email или пароль",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHASH), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "INVALID_CREDENTIALS",
			Message: "Неверный email или пароль",
		})
		return
	}

	accessToken, err := generateToken(user.ID, time.Minute*15, AccessSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "TOKEN_GENERATION_ERROR",
			Message: "Ошибка при генерации access токена",
		})
		return
	}

	refreshToken, err := generateToken(user.ID, time.Hour*24*7, refreshSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "TOKEN_GENERATION_ERROR",
			Message: "Ошибка при генерации refresh токена",
		})
		return
	}

	c.JSON(http.StatusOK, response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

}

func generateToken(userID uint, duration time.Duration, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary		Обновление access токена
// @Description	Обновление access токена с помощью refresh токена
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			refresh_token	body		RefreshTokenRequest		true	"Refresh токен"
// @Success		200				{object}	response.TokenResponse	"Успешное обновление access токена"
// @Failure		400				{object}	response.ErrorResponse	"Ошибка валидации данных (VALIDATION_ERROR)"
// @Failure		401				{object}	response.ErrorResponse	"Неверный или просроченный refresh токен (INVALID_REFRESH_TOKEN) или пользователь не найден (USER_NOT_FOUND)"
// @Failure		500				{object}	response.ErrorResponse	"Ошибка сервера (TOKEN_GENERATION_ERROR)"
// @Router			/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Ошибка валидации данных",
			Details: err.Error(),
		})
		return
	}

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return refreshSecret, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Неверный или просроченный refresh токен",
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Неверный или просроченный refresh токен",
		})
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Неверный или просроченный refresh токен",
		})
		return
	}

	userID := uint(userIDFloat)

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    "USER_NOT_FOUND",
			Message: "Пользователь не найден",
		})
		return
	}

	newAccessToken, err := generateToken(user.ID, time.Minute*15, AccessSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "TOKEN_GENERATION_ERROR",
			Message: "Ошибка при генерации access токена",
		})
		return
	}

	newRefreshToken, err := generateToken(userID, time.Hour*24*7, refreshSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "TOKEN_GENERATION_ERROR",
			Message: "Ошибка при генерации нового refresh токена",
		})
		return
	}

	c.JSON(http.StatusOK, response.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}
