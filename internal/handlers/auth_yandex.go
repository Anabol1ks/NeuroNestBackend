package handlers

import (
	"NeuroNest/internal/config"
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type YandexUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"default_email"`
	Avatar    string `json:"default_avatar_id"`
}

// @Summary      Редирект на Yandex OAuth
// @Description  Перенаправляет пользователя на страницу авторизации Yandex
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      302  "Редирект на страницу авторизации Yandex"
// @Failure      500  {object}  response.ErrorResponse  "Ошибка сервера"
// @Router       /auth/yandex/login [get]
func YandexLoginHandler(c *gin.Context) {
	url := fmt.Sprintf("https://oauth.yandex.ru/authorize?response_type=code&client_id=%s&redirect_uri=%s",
		config.YandexClientID, config.YandexRedirectURL)
	c.Redirect(http.StatusFound, url)
}

// @Summary      Callback от Yandex OAuth
// @Description  Обрабатывает callback от Yandex OAuth, получает токены и данные пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        code  query     string  true  "Код авторизации от Yandex"
// @Success      200   {object}  response.TokenResponse  "Успешная авторизация"
// @Failure      400   {object}  response.ErrorResponse  "Ошибка валидации (OAUTH_ERROR)"
// @Failure      500   {object}  response.ErrorResponse  "Ошибка сервера (OAUTH_ERROR, DB_ERROR, TOKEN_GENERATION_ERROR)"
// @Router       /auth/yandex/callback [get]
func YandexCallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    "OAUTH_ERROR",
			Message: "Не удалось получить код авторизации",
		})
		return
	}

	// Exchange code for token
	tokenResp, err := http.PostForm("https://oauth.yandex.ru/token", map[string][]string{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {config.YandexClientID},
		"client_secret": {config.YandexClientSecret},
		"redirect_uri":  {config.YandexRedirectURL},
	})
	if err != nil || tokenResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "OAUTH_ERROR",
			Message: "Ошибка при получении токена",
		})
		return
	}
	defer tokenResp.Body.Close()

	var tokenData map[string]interface{}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "OAUTH_ERROR",
			Message: "Ошибка при обработке токена",
		})
		return
	}

	accessToken := tokenData["access_token"].(string)

	// Fetch user info
	userResp, err := http.Get("https://login.yandex.ru/info?format=json&oauth_token=" + accessToken)
	if err != nil || userResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "OAUTH_ERROR",
			Message: "Ошибка при получении данных пользователя",
		})
		return
	}
	defer userResp.Body.Close()

	body, _ := ioutil.ReadAll(userResp.Body)
	var yandexUser YandexUser
	if err := json.Unmarshal(body, &yandexUser); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    "OAUTH_ERROR",
			Message: "Ошибка при обработке данных пользователя",
		})
		return
	}

	avatarUrl := fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", yandexUser.Avatar)

	// Check if user exists
	var user models.User
	if err := db.DB.Where("yandex_id = ?", yandexUser.ID).First(&user).Error; err != nil {
		// Register new user
		user = models.User{
			Nickname:   yandexUser.FirstName,
			Email:      yandexUser.Email,
			YandexID:   &yandexUser.ID,
			FirstName:  yandexUser.FirstName,
			LastName:   yandexUser.LastName,
			ProfilePic: avatarUrl,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Code:    "DB_ERROR",
				Message: "Ошибка при создании пользователя",
			})
			return
		}
	}

	accessToken, err = generateToken(user.ID, time.Minute*15, AccessSecret)
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
	frontendURL := os.Getenv("FRONT_URL")
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s", frontendURL, accessToken, refreshToken)
	c.Redirect(http.StatusFound, redirectURL)
}
