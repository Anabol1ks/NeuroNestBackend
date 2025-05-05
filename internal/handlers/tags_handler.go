package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TagInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateTagsHandler godoc
// @Security		BearerAuth
// @Summary		Создать тег
// @Description	Создаёт новый тег
// @Tags			tag
// @Accept		json
// @Produce		json
// @Param			tag body TagInput true "Данные заметки"
// @Success		201	{object}	response.SuccessResponse	"Тег успешно создан"
// @Failure		400	{object}	response.ErrorResponse	"Ошибка валидации"
// @Failure		500	{object}	response.ErrorResponse	"Ошибка при создании тега"
// @Router			/tags/create [post]
func CreateTagsHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	var input TagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Ошибка валидации данных",
			Details: err.Error(),
		})
		return
	}

	tag := models.Tag{
		UserID:      userID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := db.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при создании тега",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Тег успешно создан",
	})
}

// GetTagsHandler godoc
// @Security		BearerAuth
// @Summary		Получить теги
// @Description	Возвращает список тегов пользователя
// @Tags			tag
// @Accept		json
// @Produce		json
// @Success		200	{array}	response.TagsListResponse	"Список тегов пользователя"
// @Failure		500	{object}	response.ErrorResponse	"Ошибка при получении тегов"
// @Router			/tags/list [get]
func GetTagsHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	var tags []models.Tag
	if err := db.DB.Where("user_id = ?", userID).Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при получении тегов",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	var tagsResponse []response.TagResponse
	for _, tag := range tags {
		tagsResponse = append(tagsResponse, response.TagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	c.JSON(http.StatusOK, response.TagsListResponse{
		Tags:  tagsResponse,
		Total: len(tagsResponse),
	})
}

// DeleteTagHandler godoc
// @Security		BearerAuth
// @Summary		Удаление тега
// @Description	Удаляет тег пользователя по id
// @Tags tag
// @Accept json
// @Produce json
// @Param			id	path		uint	true	"ID тега"
// @Success 200 {object} response.SuccessResponse "Тег успешно удалён"
// @Failure 404 {object} response.ErrorResponse "Тег не найден TAG_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse "Ошибка при удалении тега DB_ERROR"
// @Router	/tags/{id} [delete]
func DeleteTagHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	tagID := c.Param("id")

	// Начинаем транзакцию для обеспечения целостности данных
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Проверяем существование тега
	var tag models.Tag
	if err := tx.Where("id = ? AND user_id = ?", tagID, userID).First(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Тег не найден",
			Code:    "TAG_NOT_FOUND",
			Details: err.Error(),
		})
		return
	}

	// Удаляем связи тега с заметками
	if err := tx.Exec("DELETE FROM note_tags WHERE tag_id = ?", tagID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении связей с тегами",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Удаляем сам тег
	if err := tx.Delete(&tag).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении тега",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при фиксации транзакции",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Тег успешно удалён",
	})
}
