package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"NeuroNest/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateNoteInput структура для создания заметки
type CreateNoteInput struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	TopicID *uint  `json:"topic_id,omitempty"`
}

// CreateNoteHandler godoc
// @Security		BearerAuth
// @Summary		Создать заметку
// @Description	Создаёт новую заметку пользователя с генерацией эмбеддинга
// @Tags			note
// @Accept			json
// @Produce		json
// @Param			note	body		CreateNoteInput	true	"Данные заметки"
// @Success		201		{object}	response.SuccessResponse	"Заметка успешно создана"
// @Failure		400		{object}	response.ErrorResponse	"Ошибка валидации"
// @Failure		500		{object}	response.ErrorResponse	"Ошибка генерации эмбеддинга EMBEDDING_ERROR   Ошибка сериализации эмбеддинга EMBEDDING_SERIALIZE_ERROR"
// @Router			/note/create [post]
func CreateNoteHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	var input CreateNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Ошибка валидации данных",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	embedding, err := service.GenerateEmbedding(input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка генерации эмбеддинга",
			Code:    "EMBEDDING_ERROR",
			Details: err.Error(),
		})
		return
	}

	// Сохраняем эмбеддинг как []byte (json)
	embeddingBytes, err := json.Marshal(embedding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка сериализации эмбеддинга",
			Code:    "EMBEDDING_SERIALIZE_ERROR",
			Details: err.Error(),
		})
		return
	}

	note := models.Note{
		UserID:    userID,
		Title:     input.Title,
		Content:   input.Content,
		Embedding: embeddingBytes,
	}
	if input.TopicID != nil {
		note.TopicID = *input.TopicID
	}

	if err := db.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при создании заметки",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Заметка успешно создана",
	})
}

// SummarizeNoteByIDHandler godoc
// @Security		BearerAuth
// @Summary		Суммаризация заметки по ID
// @Description	Генерирует краткое резюме для заметки пользователя по её ID
// @Tags			note
// @Accept			json
// @Produce		json
// @Param			id	path		uint	true	"ID заметки"
// @Success		200		{object}	response.SummarizeResponse	"Резюме успешно сгенерировано"
// @Failure		404		{object}	response.ErrorResponse	"Заметка не найдена NOTE_NOT_FOUND"
// @Failure		500		{object}	response.ErrorResponse	"Ошибка генерации резюме SUMMARY_ERROR"
// @Router			/note/{id}/summarize [post]
func SummarizeNoteByIDHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	noteID := c.Param("id")

	var note models.Note
	if err := db.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	summary, err := service.SummarizeText(note.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка генерации резюме",
			Code:    "SUMMARY_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SummarizeResponse{Summary: summary})
}
