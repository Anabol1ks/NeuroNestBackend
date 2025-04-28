package handlers

import (
	"NeuroNest/internal/config"
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"NeuroNest/internal/storage"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// UploadAvatarHandler godoc
// @Security		BearerAuth
// @Summary		Загрузка аватарки пользователя
// @Description	Позволяет пользователю загрузить аватарку. Поддерживаются форматы PNG, JPG, JPEG. Максимальный размер файла — 2MB.
// @Tags			profile
// @Accept			multipart/form-data
// @Produce		json
// @Param			avatar	formData	file	true	"Аватарка пользователя"
// @Success		200		{object}	response.UploadAvatarResponse	"Файл успешно загружен"
// @Failure		400		{object}	response.ErrorResponse		"Ошибка валидации (например, файл слишком большой или неподдерживаемый формат)"
// @Failure 404 {object} response.ErrorResponse "Пользователь не найден CODE: USER_NOT_FOUND"
// @Failure		500		{object}	response.ErrorResponse		"Ошибка сервера (например, ошибка сохранения файла или базы данных)"
// @Router			/profile/upload-avatar [post]
func UploadAvatarHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Аватарка обязательна",
			Code:    "AVATAR_REQUIRED",
		})
		return
	}
	defer file.Close()

	// Проверяем размер/тип файла
	if header.Size > 2<<20 { // 2MB
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Файл слишком большой",
			Code:    "FILE_TOO_LARGE",
		})
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Неподдерживаемый формат файла",
			Code:    "UNSUPPORTED_FORMAT",
		})
		return
	}

	filename := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Пользователь не найден",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	// Сохраняем файл через сервис
	avatarSvc := storage.NewLocalAvatarService(
		config.UploadsPath,
		config.BaseURL+"/avatars",
	)

	if user.ProfilePic != "" {
		oldFilename := filepath.Base(user.ProfilePic)
		if err := avatarSvc.Delete(oldFilename); err != nil {
			fmt.Printf("Ошибка при удалении старой аватарки: %v\n", err)
		}
	}

	avatarURL, err := avatarSvc.Save(file, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Не удалось сохранить аватарку",
			Code:    "FILE_SAVE_ERROR",
		})
		return
	}

	// Обновляем ссылку на аватарку в базе данных
	if err := db.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("profile_pic", avatarURL).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка базы данных",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, response.UploadAvatarResponse{
		Message:    "Файл успешно загружен",
		ProfilePic: avatarURL,
	})
}

// DeleteAvatarHandler godoc
// @Security		BearerAuth
// @Summary		Удаление аватарки пользователя
// @Description	Удаляет аватарку пользователя с сервера и очищает ссылку в базе данных.
// @Tags			profile
// @Accept			json
// @Produce		json
// @Success		200		{object}	response.SuccessResponse	"Аватарка успешно удалена"
// @Failure		404		{object}	response.ErrorResponse		"Аватарка не найдена"
// @Failure		500		{object}	response.ErrorResponse		"Ошибка сервера"
// @Router			/profile/delete-avatar [delete]
func DeleteAvatarHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	// Получаем пользователя из базы данных
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Пользователь не найден",
			Code:    "USER_NOT_FOUND",
		})
		return
	}

	if user.ProfilePic == "" {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Аватарка не найдена",
			Code:    "AVATAR_NOT_FOUND",
		})
		return
	}

	// Удаляем файл аватарки
	avatarSvc := storage.NewLocalAvatarService(
		config.UploadsPath,
		config.BaseURL+"/avatars",
	)
	filename := filepath.Base(user.ProfilePic)
	if err := avatarSvc.Delete(filename); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Не удалось удалить аватарку",
			Code:    "FILE_DELETE_ERROR",
		})
		return
	}

	user.ProfilePic = ""
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при обновлении профиля",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Аватарка успешно удалена",
	})
}
