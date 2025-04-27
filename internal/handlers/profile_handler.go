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
