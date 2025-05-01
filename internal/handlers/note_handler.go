package handlers

import (
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"NeuroNest/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// CreateNoteInput структура для создания заметки
type CreateNoteInput struct {
	Title      string  `json:"title" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	RelatedIDs []int64 `json:"related_ids,omitempty"`
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
// @Router			/notes/create [post]
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

	// 1) Сгенерировать embedding
	embedding, err := service.GenerateEmbedding(input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка генерации эмбеддинга",
			Code:    "EMBEDDING_ERROR",
			Details: err.Error(),
		})
		return
	}
	embeddingBytes, err := json.Marshal(embedding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка сериализации эмбеддинга",
			Code:    "EMBEDDING_SERIALIZE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 2) Собираем модель Note, прокидываем RelatedIDs (можно не проверять на len)
	note := models.Note{
		UserID:     userID,
		Title:      input.Title,
		Content:    input.Content,
		Embedding:  embeddingBytes,
		RelatedIDs: pq.Int64Array(input.RelatedIDs), // тут либо nil, либо []uint{…}
	}

	// 3) Сохраняем
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
// @Failure		500		{object}	response.ErrorResponse	"Ошибка генерации резюме SUMMARY_ERROR, Ошибка сохранения резюме SUMMARY_SAVE_ERROR"
// @Router			/notes/{id}/summarize [post]
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

	note.Summary = summary
	if err := db.DB.Model(&note).Update("summary", summary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка сохранения резюме",
			Code:    "SUMMARY_SAVE_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SummarizeResponse{Summary: summary})
}

// GetNotesHandler godoc
// @Security		BearerAuth
// @Summary		Получения списка заметок
// @Description	Выдаёт список всех заметок авторизованного пользователя
// @Tags note
// @Accept json
// @Produce json
// @Success 200 {object} response.NotesListResponse "Список заметок"
// @Failure 500 {object} response.ErrorResponse "Ошибка при получении заметок: DB_ERROR"
// @Router			/notes/list [get]
func GetNotesHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	var notes []models.Note
	if err := db.DB.Where("user_id = ?", userID).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при получении заметок",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	var notesResp []response.NoteResponse
	for _, note := range notes {
		// Преобразование вложений
		var attachments []response.AttachmentShort
		for _, att := range note.Attachments {
			attachments = append(attachments, response.AttachmentShort{
				ID:       att.ID,
				FileURL:  att.FileURL,
				FileType: att.FileType,
				FileSize: att.FileSize,
			})
		}
		// Преобразование тегов
		var tags []response.TagShort
		for _, tag := range note.Tags {
			tags = append(tags, response.TagShort{
				ID:   tag.ID,
				Name: tag.Name,
			})
		}
		notesResp = append(notesResp, response.NoteResponse{
			ID:          note.ID,
			Title:       note.Title,
			Content:     note.Content,
			Summary:     note.Summary,
			Attachments: attachments,
			IsArchived:  note.IsArchived,
			Tags:        tags,
			RelatedIDs:  note.RelatedIDs,
			CreatedAt:   note.CreatedAt.Format("2006-01-02"),
			UpdatedAt:   note.UpdatedAt.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, response.NotesListResponse{
		Notes: notesResp,
		Total: len(notesResp),
	})
}

// GetNoteHandler godoc
// @Security		BearerAuth
// @Summary		Получения заметки
// @Description	Получение заметки пользователя по id
// @Tags note
// @Accept json
// @Produce json
// @Param			id	path		uint	true	"ID заметки"
// @Success 200 {object} response.NoteResponse "Заметка успешно получена"
// @Failure 404 {object} response.ErrorResponse "Заметка не найдена NOTE_NOT_FOUND"
// @Router	/notes/{id} [get]
func GetNoteHandler(c *gin.Context) {
	noteID := c.Param("id")
	userID := c.GetUint("userID")

	var note models.Note
	if err := db.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, response.NoteResponse{
		ID:         note.ID,
		Title:      note.Title,
		Content:    note.Content,
		Summary:    note.Summary,
		IsArchived: note.IsArchived,
		RelatedIDs: note.RelatedIDs,
		CreatedAt:  note.CreatedAt.Format("2006-01-02"),
		UpdatedAt:  note.UpdatedAt.Format("2006-01-02"),
	})
}

// ArchiveNoteHandler godoc
// @Security		BearerAuth
// @Summary		Архивировать заметку
// @Description	Переносит заметку пользователя в архив (IsArchived = true)
// @Tags			note
// @Accept			json
// @Produce		json
// @Param			id	path	uint	true	"ID заметки"
// @Success		200	{object}	response.SuccessResponse	"Заметка архивирована"
// @Failure		404	{object}	response.ErrorResponse	"Заметка не найдена NOTE_NOT_FOUND"
// @Failure		500	{object}	response.ErrorResponse	"Ошибка при архивировании заметки DB_ERROR"
// @Router			/notes/{id}/archive [patch]
func ArchiveNoteHandler(c *gin.Context) {
	noteID := c.Param("id")
	userID := c.GetUint("userID")

	var note models.Note
	if err := db.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	if note.IsArchived {
		c.JSON(http.StatusOK, response.SuccessResponse{
			Message: "Заметка уже в архиве",
		})
		return
	}

	if err := db.DB.Model(&note).Update("is_archived", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при архивировании заметки",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Заметка архивирована",
	})
}

// DeleteNoteHandler godoc
// @Security		BearerAuth
// @Summary		Удаление заметки
// @Description	Удаляет заметку пользователя по id
// @Tags note
// @Accept json
// @Produce json
// @Param			id	path		uint	true	"ID заметки"
// @Success 200 {object} response.SuccessResponse "Заметка успешно удалена"
// @Failure 404 {object} response.ErrorResponse "Заметка не найдена NOTE_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse "Ошибка при удалении заметки DB_ERROR"
// @Router	/notes/{id} [delete]
func DeleteNoteHandler(c *gin.Context) {
	noteID := c.Param("id")
	userID := c.GetUint("userID")

	var note models.Note
	if err := db.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	if err := db.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении заметки",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Заметка успешно удалена",
	})
}
