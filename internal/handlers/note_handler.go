package handlers

import (
	"NeuroNest/internal/config"
	"NeuroNest/internal/db"
	"NeuroNest/internal/models"
	"NeuroNest/internal/response"
	"NeuroNest/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateNoteInput структура для создания заметки (multipart/form-data)
type CreateNoteInput struct {
	Title      string  `form:"title" binding:"required"`
	Content    string  `form:"content" binding:"required"`
	RelatedIDs []int64 `form:"related_ids[]"` // optional, IDs связанных заметок
	TagIDs     []uint  `form:"tag_ids[]"`     // optional, IDs тегов
}

// CreateNoteHandler godoc
// @Security		BearerAuth
// @Summary		Создать заметку
// @Description	Создаёт новую заметку пользователя с генерацией эмбеддинга, тегами и вложениями
// @Tags			note
// @Accept			multipart/form-data
// @Produce		json
// @Param			title			formData	string	true	"Заголовок"
// @Param			content			formData	string	true	"Содержимое"
// @Param			related_ids		formData	[]int	false	"ID связанных заметок"
// @Param			tag_ids			formData	[]int	false	"ID тегов"
// @Param			attachments	formData	[]file	false	"Вложения (image, audio, pdf)"
// @Success		201	{object}	response.SuccessResponse	"Заметка успешно создана"
// @Failure		400	{object}	response.ErrorResponse	"Ошибка валидации"
// @Failure		500	{object}	response.ErrorResponse	"Ошибка сервера"
// @Router			/notes/create [post]
func CreateNoteHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	// 1) Биндим поля формы
	var input CreateNoteInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Message: "Ошибка валидации данных",
			Code:    "VALIDATION_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 2) Генерация embedding
	embedding, err := service.GenerateEmbedding(input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка генерации эмбеддинга",
			Code:    "EMBEDDING_ERROR",
			Details: err.Error(),
		})
		return
	}
	embBytes, err := json.Marshal(embedding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка сериализации эмбеддинга",
			Code:    "EMBEDDING_SERIALIZE_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 3) Подготовка модели заметки
	note := models.Note{
		UserID:     userID,
		Title:      input.Title,
		Content:    input.Content,
		Embedding:  embBytes,
		RelatedIDs: pq.Int64Array(input.RelatedIDs),
	}

	// 4) Сохраняем заметку, чтобы получить note.ID (без тегов)
	if err := db.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при создании заметки",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 5) Если указаны теги — подгружаем их из БД и связываем
	if len(input.TagIDs) > 0 {
		// Создаем записи в промежуточной таблице note_tags
		for _, tagID := range input.TagIDs {
			if err := db.DB.Exec("INSERT INTO note_tags (note_id, tag_id) VALUES (?, ?)", note.ID, tagID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Message: "Ошибка при связывании заметки с тегами",
					Code:    "DB_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}

	// 6) Обработка файлов attachments (поле formData file, multi)
	form, err := c.MultipartForm()
	if err == nil && form.File["attachments"] != nil {
		for _, fh := range form.File["attachments"] {
			ext := strings.ToLower(filepath.Ext(fh.Filename))
			var fType string
			switch ext {
			case ".png", ".jpg", ".jpeg", ".gif":
				fType = "image"
			case ".mp3", ".wav", ".ogg":
				fType = "audio"
			case ".pdf":
				fType = "pdf"
			default:
				// пропускаем неподдерживаемый формат
				continue
			}

			// генерируем уникальное имя
			newName := fmt.Sprintf("%d_%s%s", userID, uuid.New().String(), ext)
			dst := filepath.Join(config.UploadsPath+"/attachments", newName)
			if err := c.SaveUploadedFile(fh, dst); err != nil {
				// просто логируем и продолжаем
				fmt.Printf("file save error: %v\n", err)
				continue
			}

			url := fmt.Sprintf("/attachments/%s", newName)
			att := models.Attachment{
				NoteID:     note.ID,
				FileURL:    url,
				FileType:   fType,
				FileSize:   fh.Size,
				UploadedAt: time.Now(),
			}
			db.DB.Create(&att)
		}
	}

	// 7) Ответ
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
	if err := db.DB.Where("user_id = ?", userID).Preload("Tags").Preload("Attachments").Find(&notes).Error; err != nil {
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
	if err := db.DB.Where("id = ? AND user_id = ?", noteID, userID).Preload("Tags").Preload("Attachments").First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	var attachments []response.AttachmentShort
	for _, att := range note.Attachments {
		attachments = append(attachments, response.AttachmentShort{
			ID:       att.ID,
			FileURL:  att.FileURL,
			FileType: att.FileType,
			FileSize: att.FileSize,
		})
	}

	var tags []response.TagShort
	for _, tag := range note.Tags {
		tags = append(tags, response.TagShort{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	c.JSON(http.StatusOK, response.NoteResponse{
		ID:          note.ID,
		Title:       note.Title,
		Content:     note.Content,
		Summary:     note.Summary,
		IsArchived:  note.IsArchived,
		Tags:        tags,
		Attachments: attachments,
		RelatedIDs:  note.RelatedIDs,
		CreatedAt:   note.CreatedAt.Format("2006-01-02"),
		UpdatedAt:   note.UpdatedAt.Format("2006-01-02"),
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
// @Description	Удаляет заметку пользователя по id и все связанные вложения
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

	// Начинаем транзакцию для обеспечения целостности данных
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var note models.Note
	if err := tx.Where("id = ? AND user_id = ?", noteID, userID).Preload("Attachments").First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Message: "Заметка не найдена",
			Code:    "NOTE_NOT_FOUND",
		})
		return
	}

	// 1. Удаляем связи между заметкой и тегами в промежуточной таблице
	if err := tx.Exec("DELETE FROM note_tags WHERE note_id = ?", note.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении связей с тегами",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 2. Удаляем физические файлы вложений
	for _, attachment := range note.Attachments {
		// Извлекаем имя файла из URL
		fileURL := attachment.FileURL
		fileName := filepath.Base(fileURL)

		filePath := filepath.Join(config.UploadsPath, "attachments", fileName)

		if err := os.Remove(filePath); err != nil {
			// Логируем ошибку, но продолжаем выполнение
			fmt.Printf("Error deleting file %s: %v\n", filePath, err)
		}
	}

	// 3. Удаляем вложения из базы данных
	if err := tx.Where("note_id = ?", note.ID).Delete(&models.Attachment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении вложений",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}

	// 4. Удаляем саму заметку
	if err := tx.Delete(&note).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Message: "Ошибка при удалении заметки",
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
		Message: "Заметка и все связанные данные успешно удалены",
	})
}
