package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetProfileHandler godoc
// @Security		BearerAuth
// @Summary		Получение информации о профиле
// @Description	Получает информацию о пользователе по его ID
// @Tags			profile
// @Accept			json
// @Produce		json
// @Success		200	{object}	response.ProfileResponse	"Информация о профиле пользователя"
// @Failure		404	{object}	response.ErrorResponse		"Пользователь не найден"
// @Router			/profile/get [get]
func GetProfileHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Message: "Пользователь не найден"})
		return
	}

	userRes := response.ProfileResponse{
		Nickname:   user.Nickname,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		ProfilePic: user.ProfilePic,
	}
	c.JSON(http.StatusOK, userRes)
}

// UpdateProfileHandler godoc
// @Security		BearerAuth
// @Summary		Обновление информации профиля
// @Description	Обновляет информацию профиля пользователя (кроме email)
// @Tags			profile
// @Accept			json
// @Produce		json
// @Param			profile	body		UpdateProfileInput	true	"Данные для обновления профиля"
// @Success		200		{object}	response.SuccessResponse	"Профиль успешно обновлен"
// @Failure		400		{object}	response.ErrorResponse		"Ошибка валидации данных"
// @Failure		404		{object}	response.ErrorResponse		"Пользователь не найден"
// @Router			/profile/update [put]
func UpdateProfileHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Ошибка валидации данных",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Пользователь не найден",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	// Обновляем только измененные поля
	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.ProfilePic != nil {
		user.ProfilePic = *input.ProfilePic
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при обновлении профиля",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Профиль успешно обновлен",
	})
}

// UpdateProfileInput структура для входных данных обновления профиля
type UpdateProfileInput struct {
	Nickname   *string `json:"nickname,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	ProfilePic *string `json:"profile_pic,omitempty"`
}
